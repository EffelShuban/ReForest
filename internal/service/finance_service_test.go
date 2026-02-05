package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"reforest/internal/models"
	"reforest/pkg/pb"

	"github.com/google/uuid"
)

type mockFinanceRepo struct {
	balanceUserID      uuid.UUID
	balanceDelta       int64
	getBalanceValue    int64
	getBalanceErr      error
	getEmail           string
	getEmailBalanceVal int64
	getEmailBalanceErr error
	createTx           *models.Transaction
	createTxErr        error
	txByID             *models.Transaction
	txByIDErr          error
	txsByUser          []models.Transaction
	txsByUserErr       error
	updatedStatusID    uuid.UUID
	updatedStatusValue string
	updateStatusErr    error
	invoiceUpdateID    uuid.UUID
	invoiceUpdateURL   string
	invoiceUpdateExp   time.Time
	invoiceUpdateErr   error
	pendingBefore      []models.Transaction
	pendingBeforeErr   error
}

func (m *mockFinanceRepo) GetUserBalance(ctx context.Context, userID uuid.UUID) (int64, error) {
	return m.getBalanceValue, m.getBalanceErr
}

func (m *mockFinanceRepo) UpdateWalletBalance(ctx context.Context, userID uuid.UUID, amount int64) error {
	m.balanceUserID = userID
	m.balanceDelta = amount
	return m.getBalanceErr
}

func (m *mockFinanceRepo) GetUserEmailAndBalance(ctx context.Context, userID uuid.UUID) (string, int64, error) {
	return m.getEmail, m.getEmailBalanceVal, m.getEmailBalanceErr
}

func (m *mockFinanceRepo) CreateTransaction(ctx context.Context, tx *models.Transaction) error {
	m.createTx = tx
	return m.createTxErr
}

func (m *mockFinanceRepo) GetTransactionByID(ctx context.Context, id uuid.UUID) (*models.Transaction, error) {
	return m.txByID, m.txByIDErr
}

func (m *mockFinanceRepo) GetTransactionsByUserID(ctx context.Context, userID uuid.UUID) ([]models.Transaction, error) {
	return m.txsByUser, m.txsByUserErr
}

func (m *mockFinanceRepo) UpdateTransactionStatus(ctx context.Context, id uuid.UUID, status string) error {
	m.updatedStatusID = id
	m.updatedStatusValue = status
	return m.updateStatusErr
}

func (m *mockFinanceRepo) UpdateTransactionInvoiceDetails(ctx context.Context, id uuid.UUID, paymentURL string, expiresAt time.Time) error {
	m.invoiceUpdateID = id
	m.invoiceUpdateURL = paymentURL
	m.invoiceUpdateExp = expiresAt
	return m.invoiceUpdateErr
}

func (m *mockFinanceRepo) GetPendingTransactionsBefore(ctx context.Context, expiryTime time.Time) ([]models.Transaction, error) {
	return m.pendingBefore, m.pendingBeforeErr
}

func TestFinanceService_CreateTransaction_AdoptCreatesInvoiceWhenBalanceInsufficient(t *testing.T) {
	repo := &mockFinanceRepo{getBalanceValue: 500}
	svc := NewFinanceService(repo, "key", nil, nil)

	expiry := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	body := fmt.Sprintf(`{"invoice_url":"https://pay.xendit.co/abc","status":"PENDING","expiry_date":"%s"}`, expiry.Format(time.RFC3339))

	origTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://api.xendit.co/v2/invoices" {
			t.Fatalf("unexpected URL called: %s", r.URL.String())
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})
	defer func() { http.DefaultTransport = origTransport }()

	req := &pb.TransactionRequest{
		UserId:      uuid.New().String(),
		Amount:      1500,
		Type:        pb.TransactionType_ADOPT,
		ReferenceId: "ref-1",
	}

	tx, err := svc.CreateTransaction(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateTransaction() error = %v", err)
	}

	if repo.createTx == nil {
		t.Fatalf("CreateTransaction should be called on repo")
	}
	if repo.balanceDelta != 0 {
		t.Fatalf("wallet should not be updated when balance insufficient, got delta %d", repo.balanceDelta)
	}
	if repo.invoiceUpdateURL != "https://pay.xendit.co/abc" {
		t.Fatalf("invoice URL not saved")
	}
	if !repo.invoiceUpdateExp.Equal(expiry) {
		t.Fatalf("expiry not saved correctly, got %v want %v", repo.invoiceUpdateExp, expiry)
	}
	if tx.Payment.Status != "PENDING" {
		t.Fatalf("expected payment status PENDING, got %s", tx.Payment.Status)
	}
}

func TestFinanceService_CreateTransaction_InvalidUser(t *testing.T) {
	repo := &mockFinanceRepo{}
	svc := NewFinanceService(repo, "key", nil, nil)

	_, err := svc.CreateTransaction(context.Background(), &pb.TransactionRequest{
		UserId: "invalid-uuid",
		Amount: 100,
		Type:   pb.TransactionType_ADOPT,
	})
	if err == nil {
		t.Fatalf("expected error for invalid user id")
	}
}

func TestFinanceService_CreateTransaction_GetBalanceError(t *testing.T) {
	repo := &mockFinanceRepo{getBalanceErr: errors.New("db down")}
	svc := NewFinanceService(repo, "key", nil, nil)

	_, err := svc.CreateTransaction(context.Background(), &pb.TransactionRequest{
		UserId: uuid.New().String(),
		Amount: 100,
		Type:   pb.TransactionType_ADOPT,
	})
	if err == nil {
		t.Fatalf("expected error from balance retrieval")
	}
}

func TestFinanceService_CreateTransaction_UnhandledType(t *testing.T) {
	repo := &mockFinanceRepo{}
	svc := NewFinanceService(repo, "key", nil, nil)

	_, err := svc.CreateTransaction(context.Background(), &pb.TransactionRequest{
		UserId: uuid.New().String(),
		Amount: 10,
		Type:   pb.TransactionType_CARE,
	})
	if err == nil {
		t.Fatalf("expected error for unhandled type")
	}
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestFinanceService_TopUpWallet_Success(t *testing.T) {
	repo := &mockFinanceRepo{}
	svc := NewFinanceService(repo, "xnd_key", nil, nil)

	expiry := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	body := `{"invoice_url":"https://pay.xendit.co/abc","status":"PENDING","expiry_date":"` + expiry.Format(time.RFC3339) + `"}`

	origTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.String() != "https://api.xendit.co/v2/invoices" {
			t.Fatalf("unexpected URL called: %s", r.URL.String())
		}
		return &http.Response{
			StatusCode: 200,
			Body:       io.NopCloser(strings.NewReader(body)),
			Header:     make(http.Header),
		}, nil
	})
	defer func() { http.DefaultTransport = origTransport }()

	userID := uuid.New()
	amount := int64(2000)
	tx, err := svc.TopUpWallet(context.Background(), userID, amount, 3600)
	if err != nil {
		t.Fatalf("TopUpWallet() error = %v", err)
	}

	if repo.createTx == nil || repo.createTx.UserID != userID {
		t.Fatalf("transaction should be created for user")
	}
	if repo.invoiceUpdateURL != "https://pay.xendit.co/abc" {
		t.Fatalf("invoice URL not saved")
	}
	if !repo.invoiceUpdateExp.Equal(expiry) {
		t.Fatalf("expiry not saved correctly, got %v want %v", repo.invoiceUpdateExp, expiry)
	}
	if tx.Payment.PaymentURL != repo.invoiceUpdateURL {
		t.Fatalf("payment URL not set in response")
	}
}

func TestFinanceService_TopUpWallet_XenditErrorStatus(t *testing.T) {
	repo := &mockFinanceRepo{}
	svc := NewFinanceService(repo, "x", nil, nil)

	origTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 500, Body: io.NopCloser(strings.NewReader("err")), Header: make(http.Header)}, nil
	})
	defer func() { http.DefaultTransport = origTransport }()

	_, err := svc.TopUpWallet(context.Background(), uuid.New(), 1000, 3600)
	if err == nil {
		t.Fatalf("expected error on non-200 status")
	}
}

func TestFinanceService_TopUpWallet_InvalidExpiryFallsBack(t *testing.T) {
	repo := &mockFinanceRepo{}
	svc := NewFinanceService(repo, "x", nil, nil)

	body := `{"invoice_url":"https://pay.xendit.co/abc","status":"PENDING","expiry_date":"not-a-date"}`
	now := time.Now()
	origTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	defer func() { http.DefaultTransport = origTransport }()

	tx, err := svc.TopUpWallet(context.Background(), uuid.New(), 1000, 10)
	if err != nil {
		t.Fatalf("TopUpWallet error: %v", err)
	}
	// fallback expiry should be after now
	if !tx.Payment.ExpiresAt.After(now) {
		t.Fatalf("expected fallback expiry time to be set")
	}
}

func TestFinanceService_TopUpWallet_UpdateInvoiceError(t *testing.T) {
	repo := &mockFinanceRepo{invoiceUpdateErr: errors.New("fail")}
	svc := NewFinanceService(repo, "x", nil, nil)

	body := `{"invoice_url":"https://pay.xendit.co/abc","status":"PENDING","expiry_date":"2024-01-01T00:00:00Z"}`
	origTransport := http.DefaultTransport
	http.DefaultTransport = roundTripperFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(body)), Header: make(http.Header)}, nil
	})
	defer func() { http.DefaultTransport = origTransport }()

	if _, err := svc.TopUpWallet(context.Background(), uuid.New(), 1000, 10); err == nil {
		t.Fatalf("expected error when update invoice details fails")
	}
}

func TestFinanceService_HandleWalletWebhook_Paid(t *testing.T) {
	repo := &mockFinanceRepo{}
	svc := NewFinanceService(repo, "x", nil, nil)

	txID := uuid.New()
	repo.txByID = &models.Transaction{
		ID: txID,
		Payment: models.Payment{
			Status: "PENDING",
		},
		UserID: txID,
		Amount: 900,
		Type:   "DEPOSIT",
	}

	payload := `{"external_id":"` + txID.String() + `","status":"PAID"}`
	err := svc.HandleWalletWebhook(context.Background(), "INVOICE_CALLBACK", []byte(payload))
	if err != nil {
		t.Fatalf("HandleWalletWebhook() error = %v", err)
	}

	if repo.updatedStatusValue != "SUCCESS" {
		t.Fatalf("expected status updated to SUCCESS, got %s", repo.updatedStatusValue)
	}
	if repo.balanceDelta != 900 {
		t.Fatalf("wallet not updated correctly, delta %d", repo.balanceDelta)
	}
}

func TestFinanceService_HandleWalletWebhook_InvalidJSON(t *testing.T) {
	repo := &mockFinanceRepo{}
	svc := NewFinanceService(repo, "x", nil, nil)
	if err := svc.HandleWalletWebhook(context.Background(), "INVOICE_CALLBACK", []byte("{bad json")); err == nil {
		t.Fatalf("expected parse error")
	}
}

func TestFinanceService_HandleWalletWebhook_InvalidExternalID(t *testing.T) {
	repo := &mockFinanceRepo{}
	svc := NewFinanceService(repo, "x", nil, nil)
	if err := svc.HandleWalletWebhook(context.Background(), "INVOICE_CALLBACK", []byte(`{"external_id":"bad","status":"PAID"}`)); err == nil {
		t.Fatalf("expected invalid uuid error")
	}
}

func TestFinanceService_HandleWalletWebhook_AlreadySuccess(t *testing.T) {
	repo := &mockFinanceRepo{
		txByID: &models.Transaction{
			ID:      uuid.New(),
			Payment: models.Payment{Status: "SUCCESS"},
		},
	}
	svc := NewFinanceService(repo, "x", nil, nil)
	payload := `{"external_id":"` + repo.txByID.ID.String() + `","status":"PAID"}`
	if err := svc.HandleWalletWebhook(context.Background(), "INVOICE_CALLBACK", []byte(payload)); err != nil {
		t.Fatalf("expected nil when already success")
	}
}

func TestFinanceService_HandleWalletWebhook_Expired(t *testing.T) {
	repo := &mockFinanceRepo{}
	svc := NewFinanceService(repo, "x", nil, nil)

	txID := uuid.New()
	repo.txByID = &models.Transaction{ID: txID, Payment: models.Payment{Status: "PENDING"}, Type: "DEPOSIT"}

	payload := `{"external_id":"` + txID.String() + `","status":"EXPIRED"}`
	err := svc.HandleWalletWebhook(context.Background(), "INVOICE_CALLBACK", []byte(payload))
	if err != nil {
		t.Fatalf("HandleWalletWebhook() error = %v", err)
	}
	if repo.updatedStatusValue != "EXPIRED" {
		t.Fatalf("status should be marked expired")
	}
}

func TestFinanceService_HandleWalletWebhook_UnknownStatus(t *testing.T) {
	repo := &mockFinanceRepo{}
	svc := NewFinanceService(repo, "x", nil, nil)
	txID := uuid.New()
	payload := `{"external_id":"` + txID.String() + `","status":"PENDING"}`
	if err := svc.HandleWalletWebhook(context.Background(), "INVOICE_CALLBACK", []byte(payload)); err != nil {
		t.Fatalf("expected nil for unknown status, got %v", err)
	}
}

func TestFinanceService_CheckPaymentExpiry_Error(t *testing.T) {
	repo := &mockFinanceRepo{pendingBeforeErr: errors.New("db")}
	svc := NewFinanceService(repo, "x", nil, nil)
	if err := svc.CheckPaymentExpiry(context.Background()); err == nil {
		t.Fatalf("expected error from repo")
	}
}

func TestFinanceService_CheckPaymentExpiry(t *testing.T) {
	repo := &mockFinanceRepo{
		pendingBefore: []models.Transaction{
			{ID: uuid.New(), Type: "DEPOSIT"},
			{ID: uuid.New(), Type: "DEPOSIT"},
		},
	}
	svc := NewFinanceService(repo, "x", nil, nil)

	if err := svc.CheckPaymentExpiry(context.Background()); err != nil {
		t.Fatalf("CheckPaymentExpiry() error = %v", err)
	}
	if repo.updatedStatusID == uuid.Nil {
		t.Fatalf("UpdateTransactionStatus should be called at least once")
	}
}

func TestFinanceService_GetBalance(t *testing.T) {
	repo := &mockFinanceRepo{getBalanceValue: 1234}
	svc := NewFinanceService(repo, "x", nil, nil)

	uid := uuid.New()
	bal, err := svc.GetBalance(context.Background(), uid)
	if err != nil {
		t.Fatalf("GetBalance() error = %v", err)
	}
	if bal != 1234 {
		t.Fatalf("balance mismatch, got %v", bal)
	}
}

func TestFinanceService_GetTransactionHistory(t *testing.T) {
	repo := &mockFinanceRepo{
		txsByUser: []models.Transaction{
			{ID: uuid.New(), Amount: 10},
			{ID: uuid.New(), Amount: 20},
		},
	}
	svc := NewFinanceService(repo, "x", nil, nil)

	uid := uuid.New()
	txs, err := svc.GetTransactionHistory(context.Background(), uid)
	if err != nil {
		t.Fatalf("GetTransactionHistory() error = %v", err)
	}
	if len(txs) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(txs))
	}
}
