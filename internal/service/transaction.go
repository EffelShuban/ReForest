package service

import (
	"context"
	"strings"

	"reforest/internal/models"
	"reforest/internal/repository"
)

type TransactionService interface {
	CreateTransaction(ctx context.Context, tx *models.Transaction) (*models.Transaction, error)
	GetTransaction(ctx context.Context, id string) (*models.Transaction, error)
	ListTransaction(ctx context.Context) ([]*models.Transaction, error)
	UpdateTransaction(ctx context.Context, id string, tx *models.Transaction) (*models.Transaction, error)
	DeleteTransaction(ctx context.Context, id string) error

	CreatePayment(ctx context.Context, payment *models.Payment) (*models.Payment, error)
	GetPayment(ctx context.Context, id string) (*models.Payment, error)
	ListPayment(ctx context.Context) ([]*models.Payment, error)
	UpdatePayment(ctx context.Context, id string, payment *models.Payment) (*models.Payment, error)
	DeletePayment(ctx context.Context, id string) error
}

type transactionService struct {
	repo repository.TransactionRepository
}

func NewTransactionService(repo repository.TransactionRepository) TransactionService {
	return &transactionService{repo: repo}
}

func (s *transactionService) CreateTransaction(ctx context.Context, tx *models.Transaction) (*models.Transaction, error) {
	return s.repo.CreateTransaction(ctx, tx)
}

func (s *transactionService) GetTransaction(ctx context.Context, id string) (*models.Transaction, error) {
	return s.repo.GetTransaction(ctx, id)
}

func (s *transactionService) ListTransaction(ctx context.Context) ([]*models.Transaction, error) {
	return s.repo.ListTransaction(ctx)
}

func (s *transactionService) UpdateTransaction(ctx context.Context, id string, tx *models.Transaction) (*models.Transaction, error) {
	tx.ID = id
	return s.repo.UpdateTransaction(ctx, tx)
}

func (s *transactionService) DeleteTransaction(ctx context.Context, id string) error {
	return s.repo.DeleteTransaction(ctx, id)
}

func (s *transactionService) CreatePayment(ctx context.Context, payment *models.Payment) (*models.Payment, error) {
	if err := validatePayment(payment); err != nil {
		return nil, err
	}
	return s.repo.CreatePayment(ctx, payment)
}

func (s *transactionService) GetPayment(ctx context.Context, id string) (*models.Payment, error) {
	return s.repo.GetPayment(ctx, id)
}

func (s *transactionService) ListPayment(ctx context.Context) ([]*models.Payment, error) {
	return s.repo.ListPayment(ctx)
}

func (s *transactionService) UpdatePayment(ctx context.Context, id string, payment *models.Payment) (*models.Payment, error) {
	payment.ID = id
	if err := validatePayment(payment); err != nil {
		return nil, err
	}
	return s.repo.UpdatePayment(ctx, payment)
}

func (s *transactionService) DeletePayment(ctx context.Context, id string) error {
	return s.repo.DeletePayment(ctx, id)
}

var (
	allowedPaymentStatus = map[string]bool{"pending": true, "paid": true, "expired": true}
)

func validatePayment(p *models.Payment) error {
	if p == nil {
		return models.ErrInvalidInput
	}
	if p.TransactionID == "" {
		return models.ErrInvalidInput
	}
	if p.Amount <= 0 {
		return models.ErrInvalidInput
	}
	p.Status = strings.ToLower(p.Status)
	if p.Status != "" && !allowedPaymentStatus[p.Status] {
		return models.ErrInvalidInput
	}
	return nil
}
