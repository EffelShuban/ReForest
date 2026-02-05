package repository

import (
	"context"
	"reforest/internal/models"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FinanceRepository interface {
	GetUserBalance(ctx context.Context, userID uuid.UUID) (int64, error)
	UpdateWalletBalance(ctx context.Context, userID uuid.UUID, amount int64) error
	GetUserEmailAndBalance(ctx context.Context, userID uuid.UUID) (string, int64, error)

	CreateTransaction(ctx context.Context, tx *models.Transaction) error
	GetTransactionByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error)
	GetTransactionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error)
	UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateTransactionInvoiceDetails(ctx context.Context, id uuid.UUID, paymentURL string, expiresAt time.Time) error
	GetPendingTransactionsBefore(ctx context.Context, expiryTime time.Time) ([]models.Transaction, error)
}

type financeRepository struct {
	db *gorm.DB
}

func NewFinanceRepository(db *gorm.DB) FinanceRepository {
	return &financeRepository{db: db}
}

func (r *financeRepository) GetUserBalance(ctx context.Context, userID uuid.UUID) (int64, error) {
	var profile models.Profile
	err := r.db.WithContext(ctx).Select("balance").Where("id = ?", userID).First(&profile).Error
	return profile.Balance, err
}

func (r *financeRepository) GetUserEmailAndBalance(ctx context.Context, userID uuid.UUID) (string, int64, error) {
	var res struct {
		Email   string
		Balance int64
	}
	err := r.db.WithContext(ctx).
		Table("users").
		Select("users.email, profiles.balance").
		Joins("JOIN profiles ON profiles.id = users.id").
		Where("users.id = ?", userID).
		Scan(&res).Error
	return res.Email, res.Balance, err
}

func (r *financeRepository) UpdateWalletBalance(ctx context.Context, userID uuid.UUID, amount int64) error {
	return r.db.WithContext(ctx).Model(&models.Profile{}).
		Where("id = ?", userID).
		Update("balance", gorm.Expr("balance + ?", amount)).Error
}

func (r *financeRepository) CreateTransaction(ctx context.Context, tx *models.Transaction) error {
	return r.db.WithContext(ctx).Create(tx).Error
}

func (r *financeRepository) GetTransactionByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.WithContext(ctx).Preload("Payment").First(&tx, "id = ?", id).Error
	return &tx, err
}

func (r *financeRepository) GetTransactionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error) {
	var txs []models.Transaction
	err := r.db.WithContext(ctx).Preload("Payment").Where("user_id = ?", userID).Order("created_at desc").Find(&txs).Error
	return txs, err
}

func (r *financeRepository) UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status string) error {
	return r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("transaction_id = ?", id).
		Update("status", status).Error
}

func (r *financeRepository) UpdateTransactionInvoiceDetails(ctx context.Context, id uuid.UUID, paymentURL string, expiresAt time.Time) error {
	return r.db.WithContext(ctx).Model(&models.Payment{}).
		Where("transaction_id = ?", id).
		Updates(map[string]interface{}{
			"payment_url": paymentURL,
			"expires_at":  expiresAt,
		}).Error
}

func (r *financeRepository) GetPendingTransactionsBefore(ctx context.Context, expiryTime time.Time) ([]models.Transaction, error) {
	var txs []models.Transaction
	err := r.db.WithContext(ctx).
		Joins("JOIN payments ON payments.transaction_id = transactions.id").
		Where("payments.status = ? AND payments.expires_at < ?", "PENDING", expiryTime).
		Find(&txs).Error
	return txs, err
}
