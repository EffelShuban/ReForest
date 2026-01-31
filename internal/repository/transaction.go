package repository

import (
	"context"
	"errors"
	"reforest/internal/models"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	transactionCollection = "transactions"
	paymentCollection     = "payments"
)

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, transaction *models.Transaction) (*models.Transaction, error)
	GetTransaction(ctx context.Context, id string) (*models.Transaction, error)
	ListTransaction(ctx context.Context) ([]*models.Transaction, error)
	UpdateTransaction(ctx context.Context, transaction *models.Transaction) (*models.Transaction, error)
	DeleteTransaction(ctx context.Context, id string) error

	CreatePayment(ctx context.Context, payment *models.Payment) (*models.Payment, error)
	GetPayment(ctx context.Context, id string) (*models.Payment, error)
	ListPayment(ctx context.Context) ([]*models.Payment, error)
	UpdatePayment(ctx context.Context, payment *models.Payment) (*models.Payment, error)
	DeletePayment(ctx context.Context, id string) error
}

type transactionRepository struct {
	db *mongo.Database
}

func NewTransactionRepository(db *mongo.Database) TransactionRepository {
	return &transactionRepository{db: db}
}

func generateID(id string) string {
	if id != "" {
		return id
	}
	return uuid.NewString()
}

func (r *transactionRepository) CreateTransaction(ctx context.Context, transaction *models.Transaction) (*models.Transaction, error) {
	transaction.ID = generateID(transaction.ID)
	if transaction.TransactionDate.IsZero() {
		transaction.TransactionDate = time.Now()
	}
	_, err := r.db.Collection(transactionCollection).InsertOne(ctx, transaction)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrAlreadyExists
		}
		return nil, err
	}
	return transaction, nil
}

func (r *transactionRepository) GetTransaction(ctx context.Context, id string) (*models.Transaction, error) {
	var tx models.Transaction
	err := r.db.Collection(transactionCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&tx)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		return nil, err
	}
	return &tx, nil
}

func (r *transactionRepository) ListTransaction(ctx context.Context) ([]*models.Transaction, error) {
	cursor, err := r.db.Collection(transactionCollection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var transactions []*models.Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		return nil, err
	}
	return transactions, nil
}

func (r *transactionRepository) UpdateTransaction(ctx context.Context, transaction *models.Transaction) (*models.Transaction, error) {
	update := bson.M{
		"$set": bson.M{
			"sponsor_id":       transaction.SponsorID,
			"adopted_tree_id":  transaction.AdoptedTreeID,
			"amount":           transaction.Amount,
			"type":             transaction.Type,
			"payment_status":   transaction.PaymentStatus,
			"transaction_date": transaction.TransactionDate,
		},
	}
	res, err := r.db.Collection(transactionCollection).UpdateOne(ctx, bson.M{"_id": transaction.ID}, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrAlreadyExists
		}
		return nil, err
	}
	if res.MatchedCount == 0 {
		return nil, models.ErrNotFound
	}
	return r.GetTransaction(ctx, transaction.ID)
}

func (r *transactionRepository) DeleteTransaction(ctx context.Context, id string) error {
	res, err := r.db.Collection(transactionCollection).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return models.ErrNotFound
	}
	return nil
}

func (r *transactionRepository) CreatePayment(ctx context.Context, payment *models.Payment) (*models.Payment, error) {
	payment.ID = generateID(payment.ID)
	if payment.CreatedAt.IsZero() {
		payment.CreatedAt = time.Now()
	}
	_, err := r.db.Collection(paymentCollection).InsertOne(ctx, payment)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrAlreadyExists
		}
		return nil, err
	}
	return payment, nil
}

func (r *transactionRepository) GetPayment(ctx context.Context, id string) (*models.Payment, error) {
	var payment models.Payment
	err := r.db.Collection(paymentCollection).FindOne(ctx, bson.M{"_id": id}).Decode(&payment)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, models.ErrNotFound
		}
		return nil, err
	}
	return &payment, nil
}

func (r *transactionRepository) ListPayment(ctx context.Context) ([]*models.Payment, error) {
	cursor, err := r.db.Collection(paymentCollection).Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var payments []*models.Payment
	if err = cursor.All(ctx, &payments); err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *transactionRepository) UpdatePayment(ctx context.Context, payment *models.Payment) (*models.Payment, error) {
	update := bson.M{
		"$set": bson.M{
			"transaction_id": payment.TransactionID,
			"amount":         payment.Amount,
			"status":         payment.Status,
			"created_at":     payment.CreatedAt,
		},
	}
	res, err := r.db.Collection(paymentCollection).UpdateOne(ctx, bson.M{"_id": payment.ID}, update)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, models.ErrAlreadyExists
		}
		return nil, err
	}
	if res.MatchedCount == 0 {
		return nil, models.ErrNotFound
	}
	return r.GetPayment(ctx, payment.ID)
}

func (r *transactionRepository) DeletePayment(ctx context.Context, id string) error {
	res, err := r.db.Collection(paymentCollection).DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		return err
	}
	if res.DeletedCount == 0 {
		return models.ErrNotFound
	}
	return nil
}
