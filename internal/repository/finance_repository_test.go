package repository

import (
	"context"
	"regexp"
	"testing"
	"time"

	"reforest/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	dbSQL, mock, err := sqlmock.New()
	assert.NoError(t, err)

	gormDB, err := gorm.Open(postgres.New(postgres.Config{
		Conn: dbSQL,
	}), &gorm.Config{})
	assert.NoError(t, err)

	return gormDB, mock
}

func TestGetUserBalance(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewFinanceRepository(db)
	userID := uuid.New()

	t.Run("success get balance", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "balance"}).
			AddRow(userID, int64(50000))

		mock.ExpectQuery(regexp.QuoteMeta(`SELECT "balance" FROM "profiles" WHERE id = $1 ORDER BY "profiles"."id" LIMIT $2`)).
			WithArgs(userID, 1).
			WillReturnRows(rows)

		balance, err := repo.GetUserBalance(context.Background(), userID)

		assert.NoError(t, err)
		assert.Equal(t, int64(50000), balance)
	})
}

func TestUpdateWalletBalance(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewFinanceRepository(db)
	userID := uuid.New()
	amount := int64(10000)

	t.Run("success update balance with expression", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "profiles" SET "balance"=balance + $1 WHERE id = $2`)).
			WithArgs(amount, userID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateWalletBalance(context.Background(), userID, amount)
		assert.NoError(t, err)
	})
}

func TestCreateTransaction(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewFinanceRepository(db)

	tx := &models.Transaction{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		Amount:      25000,
		ReferenceID: "",
		Type:        "",
	}

	t.Run("success create tx", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "transactions" ("user_id","amount","reference_id","type","created_at","updated_at","id") VALUES ($1,$2,$3,$4,$5,$6,$7) RETURNING "id"`)).
			WithArgs(tx.UserID, tx.Amount, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), tx.ID).
			WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(tx.ID))
		mock.ExpectCommit()

		err := repo.CreateTransaction(context.Background(), tx)
		assert.NoError(t, err)
	})
}

func TestGetUserEmailAndBalance(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewFinanceRepository(db)
	userID := uuid.New()

	rows := sqlmock.NewRows([]string{"email", "balance"}).
		AddRow("user@example.com", int64(15000))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT users.email, profiles.balance FROM "users" JOIN profiles ON profiles.id = users.id WHERE users.id = $1`)).
		WithArgs(userID).
		WillReturnRows(rows)

	email, balance, err := repo.GetUserEmailAndBalance(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", email)
	assert.Equal(t, int64(15000), balance)
}

func TestGetTransactionByID(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewFinanceRepository(db)
	now := time.Now()

	tx := models.Transaction{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		Amount:    5000,
		Type:      "DEPOSIT",
		CreatedAt: now,
		UpdatedAt: now,
	}

	txRows := sqlmock.NewRows([]string{"id", "user_id", "amount", "reference_id", "type", "created_at", "updated_at"}).
		AddRow(tx.ID, tx.UserID, tx.Amount, "", tx.Type, tx.CreatedAt, tx.UpdatedAt)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "transactions" WHERE id = $1 ORDER BY "transactions"."id" LIMIT $2`)).
		WithArgs(tx.ID, 1).
		WillReturnRows(txRows)

	paymentRows := sqlmock.NewRows([]string{"id", "transaction_id", "amount", "status", "external_id", "payment_url", "expires_at", "created_at", "updated_at"}).
		AddRow(uuid.New(), tx.ID, tx.Amount, "PENDING", "", "", now.Add(24*time.Hour), now, now)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payments" WHERE "payments"."transaction_id" = $1`)).
		WithArgs(tx.ID).
		WillReturnRows(paymentRows)

	got, err := repo.GetTransactionByID(context.Background(), tx.ID)
	assert.NoError(t, err)
	assert.Equal(t, tx.ID, got.ID)
	assert.Equal(t, tx.Amount, got.Payment.Amount)
}

func TestGetTransactionsByUserID(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewFinanceRepository(db)
	userID := uuid.New()
	now := time.Now()

	txID := uuid.New()
	txRows := sqlmock.NewRows([]string{"id", "user_id", "amount", "reference_id", "type", "created_at", "updated_at"}).
		AddRow(txID, userID, int64(7000), "", "DEPOSIT", now, now)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "transactions" WHERE user_id = $1 ORDER BY created_at desc`)).
		WithArgs(userID).
		WillReturnRows(txRows)

	paymentRows := sqlmock.NewRows([]string{"id", "transaction_id", "amount", "status", "external_id", "payment_url", "expires_at", "created_at", "updated_at"}).
		AddRow(uuid.New(), txID, int64(7000), "COMPLETED", "", "", now, now, now)
	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "payments" WHERE "payments"."transaction_id" = $1`)).
		WithArgs(txID).
		WillReturnRows(paymentRows)

	list, err := repo.GetTransactionsByUserID(context.Background(), userID)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, userID, list[0].UserID)
	assert.Equal(t, "COMPLETED", list[0].Payment.Status)
}

func TestUpdateTransactionInvoiceDetails(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewFinanceRepository(db)
	txID := uuid.New()
	expires := time.Now().Add(1 * time.Hour)
	paymentURL := "https://pay.example.com/invoice"

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE "payments" SET "expires_at"=$1,"payment_url"=$2,"updated_at"=$3 WHERE transaction_id = $4`)).
		WithArgs(expires, paymentURL, sqlmock.AnyArg(), txID).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := repo.UpdateTransactionInvoiceDetails(context.Background(), txID, paymentURL, expires)
	assert.NoError(t, err)
}

func TestGetPendingTransactionsBefore(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewFinanceRepository(db)
	expiry := time.Now()

	txRows := sqlmock.NewRows([]string{"id", "user_id", "amount", "reference_id", "type", "created_at", "updated_at"}).
		AddRow(uuid.New(), uuid.New(), int64(9000), "", "DEPOSIT", expiry.Add(-time.Hour), expiry.Add(-time.Hour))

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT "transactions"."id","transactions"."user_id","transactions"."amount","transactions"."reference_id","transactions"."type","transactions"."created_at","transactions"."updated_at" FROM "transactions" JOIN payments ON payments.transaction_id = transactions.id WHERE payments.status = $1 AND payments.expires_at < $2`)).
		WithArgs("PENDING", expiry).
		WillReturnRows(txRows)

	list, err := repo.GetPendingTransactionsBefore(context.Background(), expiry)
	assert.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestUpdateTransactionStatus(t *testing.T) {
	db, mock := setupMockDB(t)
	repo := NewFinanceRepository(db)
	txID := uuid.New()
	status := "COMPLETED"

	t.Run("success update status", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec(regexp.QuoteMeta(`UPDATE "payments" SET "status"=$1,"updated_at"=$2 WHERE transaction_id = $3`)).
			WithArgs(status, sqlmock.AnyArg(), txID).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		err := repo.UpdateTransactionStatus(context.Background(), txID, status)
		assert.NoError(t, err)
	})
}
