package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reforest/internal/models"
	"reforest/internal/repository"
	"reforest/pkg/mq"
	"reforest/pkg/pb"
	"time"

	"github.com/google/uuid"
)

type FinanceService interface {
	CreateTransaction(ctx context.Context, req *pb.TransactionRequest) (*models.Transaction, error)
	TopUpWallet(ctx context.Context, userID uuid.UUID, amount int64, duration int32) (*models.Transaction, error)
	HandleWalletWebhook(ctx context.Context, event string, data []byte) error
	GetBalance(ctx context.Context, userID uuid.UUID) (int64, error)
	GetTransactionHistory(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error)
	CheckPaymentExpiry(ctx context.Context) error
}

type financeService struct {
	repo         repository.FinanceRepository
	xenditAPIKey string
	mqClient     *mq.Client
}

type xenditInvoiceRequest struct {
	ExternalID      string  `json:"external_id"`
	Amount          float64 `json:"amount"`
	Description     string  `json:"description"`
	InvoiceDuration int     `json:"invoice_duration"`
	Currency        string  `json:"currency"`
}

type xenditInvoiceResponse struct {
	InvoiceURL string `json:"invoice_url"`
	Status     string `json:"status"`
	ExpiryDate string `json:"expiry_date"`
}

func NewFinanceService(repo repository.FinanceRepository, xenditAPIKey string, mqClient *mq.Client) FinanceService {
	return &financeService{
		repo:         repo,
		xenditAPIKey: xenditAPIKey,
		mqClient:     mqClient,
	}
}

func (s *financeService) CreateTransaction(ctx context.Context, req *pb.TransactionRequest) (*models.Transaction, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, errors.New("invalid user id")
	}

	switch req.Type {
	case pb.TransactionType_ADOPT:
		balance, err := s.repo.GetUserBalance(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to get user balance: %w", err)
		}

		log.Printf("[CreateTransaction] User: %s, Balance: %d, Required: %d", userID, balance, req.Amount)

		if balance >= req.Amount {
			tx := &models.Transaction{
				UserID:      userID,
				Amount:      req.Amount,
				Type:        req.Type.String(),
				ReferenceID: req.ReferenceId,
				Payment: models.Payment{
					Amount: req.Amount,
					Status: "SUCCESS",
				},
			}
			if err := s.repo.CreateTransaction(ctx, tx); err != nil {
				return nil, err
			}
			if err := s.repo.UpdateWalletBalance(ctx, userID, -req.Amount); err != nil {
				return nil, fmt.Errorf("failed to update wallet balance for adoption: %w", err)
			}

			// Publish success event immediately
			_ = s.mqClient.Publish(ctx, "payment.success", map[string]string{"reference_id": tx.ReferenceID})

			return tx, nil
		}
		return s.createInvoiceTransaction(ctx, userID, req.Amount, "ADOPT", req.ReferenceId, 86400)

	case pb.TransactionType_CARE:
		// TODO: Implement care if time suffice
		return nil, errors.New("CARE transaction type not implemented")
	}

	return nil, fmt.Errorf("unhandled transaction type: %s", req.Type.String())
}

func (s *financeService) TopUpWallet(ctx context.Context, userID uuid.UUID, amount int64, duration int32) (*models.Transaction, error) {
	return s.createInvoiceTransaction(ctx, userID, amount, "DEPOSIT", "", int(duration))
}

func (s *financeService) createInvoiceTransaction(ctx context.Context, userID uuid.UUID, amount int64, txType string, refID string, duration int) (*models.Transaction, error) {
	invoiceDuration := int(duration)
	if invoiceDuration <= 0 {
		invoiceDuration = 86400 // equals 24 hours
	}

	tx := &models.Transaction{
		UserID:    userID,
		Amount:    amount,
		Type:      txType,
		ReferenceID: refID,
		Payment: models.Payment{
			Amount:    amount,
			Status:    "PENDING",
			ExpiresAt: time.Now().Add(time.Duration(invoiceDuration) * time.Second),
		},
	}

	if err := s.repo.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	reqBody := xenditInvoiceRequest{
		ExternalID:      tx.ID.String(),
		Amount:          float64(amount),
		Description:     fmt.Sprintf("%s - User %s", txType, userID.String()),
		InvoiceDuration: invoiceDuration,
		Currency:        "IDR",
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.xendit.co/v2/invoices", bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(s.xenditAPIKey, "")
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to create xendit invoice, status: %d", resp.StatusCode)
	}

	var xenditResp xenditInvoiceResponse
	if err := json.NewDecoder(resp.Body).Decode(&xenditResp); err != nil {
		return nil, err
	}

	expiryTime, err := time.Parse(time.RFC3339, xenditResp.ExpiryDate)
	if err != nil {
		fmt.Printf("failed to parse expiry date from xendit: %v\n", err)
		expiryTime = time.Now().Add(time.Duration(invoiceDuration) * time.Second)
	}

	if err := s.repo.UpdateTransactionInvoiceDetails(ctx, tx.ID, xenditResp.InvoiceURL, expiryTime); err != nil {
		fmt.Printf("failed to update transaction invoice details: %v\n", err)
		return nil, err
	}
	tx.Payment.PaymentURL = xenditResp.InvoiceURL
	tx.Payment.ExpiresAt = expiryTime
	return tx, nil
}

func (s *financeService) HandleWalletWebhook(ctx context.Context, event string, data []byte) error {
	var payload struct {
		ExternalID string `json:"external_id"`
		Status     string `json:"status"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("failed to parse webhook data: %w", err)
	}

	if payload.ExternalID == "" {
		return errors.New("external_id is required")
	}

	txID, err := uuid.Parse(payload.ExternalID)
	if err != nil {
		return fmt.Errorf("invalid external_id format: %w", err)
	}

	if payload.Status == "PAID" || payload.Status == "SETTLED" {
		tx, err := s.repo.GetTransactionByID(ctx, txID)
		if err != nil {
			return err
		}

		if tx.Payment.Status == "SUCCESS" {
			return nil
		}

		if err := s.repo.UpdateTransactionStatus(ctx, txID, "SUCCESS"); err != nil {
			return err
		}

		if tx.Type == "ADOPT" {
			_ = s.mqClient.Publish(ctx, "payment.success", map[string]string{"reference_id": tx.ReferenceID})
			return nil
		}

		return s.repo.UpdateWalletBalance(ctx, tx.UserID, tx.Amount)
	} else if payload.Status == "EXPIRED" {
		return s.repo.UpdateTransactionStatus(ctx, txID, "EXPIRED")
	}

	return nil
}

func (s *financeService) GetBalance(ctx context.Context, userID uuid.UUID) (int64, error) {
	return s.repo.GetUserBalance(ctx, userID)
}

func (s *financeService) GetTransactionHistory(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error) {
	return s.repo.GetTransactionsByUserID(ctx, userID)
}

func (s *financeService) CheckPaymentExpiry(ctx context.Context) error {
	txs, err := s.repo.GetPendingTransactionsBefore(ctx, time.Now())
	if err != nil {
		return err
	}

	for _, tx := range txs {
		if err := s.repo.UpdateTransactionStatus(ctx, tx.ID, "EXPIRED"); err != nil {
			// TODO: Log error
			continue
		}
		if tx.Type == "ADOPT" && tx.ReferenceID != "" {
			_ = s.mqClient.Publish(ctx, "payment.expired", map[string]string{"reference_id": tx.ReferenceID})
		}
	}
	return nil
}