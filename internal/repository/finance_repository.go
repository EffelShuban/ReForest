package repository

import (

	
	// "github.com/google/uuid"
	"gorm.io/gorm"
)

// GUNAKAN UUID UNTUK ID PRIMARY KEY, DIDAPAT DARI LIBRARY YANG DI COMMENT DI ATAS

type FinanceRepository interface {
	// GetWalletByUserID(ctx context.Context, userID uuid.UUID) (*models.Wallet, error)
	// CreateWallet(ctx context.Context, wallet *models.Wallet) error
	// UpdateWalletBalance(ctx context.Context, userID uuid.UUID, amount int64) error
	
	// CreateTransaction(ctx context.Context, tx *models.Transaction) error
	// GetTransactionByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error)
	// GetTransactionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error)
	// UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status string) error
	// GetPendingTransactionsBefore(ctx context.Context, expiryTime time.Time) ([]models.Transaction, error)
}

type financeRepository struct {
	db *gorm.DB
}

func NewFinanceRepository(db *gorm.DB) FinanceRepository {
	return &financeRepository{db: db}
}