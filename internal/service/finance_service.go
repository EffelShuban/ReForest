package service

import (
	"reforest/internal/repository"
	// "reforest/pkg/pb"

	// "github.com/google/uuid"
	"github.com/xendit/xendit-go/v6"
)

type FinanceService interface {
	// CreateTransaction(ctx context.Context, req *pb.TransactionRequest) (*models.Transaction, error)
	// TopUpWallet(ctx context.Context, userID string, amount int64) (string, error)
	// HandleWalletWebhook(ctx context.Context, req *pb.WebhookRequest) error
	// GetBalance(ctx context.Context, userID string) (float64, error)
	// GetTransactionHistory(ctx context.Context, userID string) ([]models.Transaction, error)
	// CheckPaymentExpiry(ctx context.Context) error
}

type financeService struct {
	repo repository.FinanceRepository
	xenditClient *xendit.APIClient
}

func NewFinanceService(repo repository.FinanceRepository, xenditAPIKey string) FinanceService {
	client := xendit.NewClient(xenditAPIKey)
	return &financeService{repo: repo, xenditClient: client}
}

// KALO MAU, BISA IMPLEMENTASI CEK WALLET YA MBA UNTUK ERROR HANDLING KETIKA MEMBUAT TRANSACTION