package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-path/internal/domain"
)

type fakeTransactionService struct {
	creditFn   func(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error)
	debitFn    func(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error)
	transferFn func(ctx context.Context, fromUserID, toUserID int64, amount float64) (*domain.Transaction, error)
}

func (f *fakeTransactionService) Credit(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error) {
	if f.creditFn == nil {
		return nil, nil
	}
	return f.creditFn(ctx, userID, amount)
}

func (f *fakeTransactionService) Debit(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error) {
	if f.debitFn == nil {
		return nil, nil
	}
	return f.debitFn(ctx, userID, amount)
}

func (f *fakeTransactionService) Transfer(ctx context.Context, fromUserID, toUserID int64, amount float64) (*domain.Transaction, error) {
	if f.transferFn == nil {
		return nil, nil
	}
	return f.transferFn(ctx, fromUserID, toUserID, amount)
}

func (f *fakeTransactionService) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	return nil, nil
}

func (f *fakeTransactionService) GetByUserID(ctx context.Context, userID int64) ([]*domain.Transaction, error) {
	return nil, nil
}

func TestTransactionHandlerCredit_Success(t *testing.T) {
	handler := NewTransactionHandler(&fakeTransactionService{
		creditFn: func(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:        1,
				ToUserID:  userID,
				Amount:    amount,
				Type:      domain.TransactionTypeCredit,
				Status:    domain.TransactionStatusCompleted,
				CreatedAt: time.Now(),
			}, nil
		},
	})

	body := creditTransactionRequest{
		UserID: 1,
		Amount: 100,
	}

	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/credit", bytes.NewBuffer(payload))
	rec := httptest.NewRecorder()

	handler.Credit(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestTransactionHandlerCredit_InvalidMethod(t *testing.T) {
	handler := NewTransactionHandler(&fakeTransactionService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/credit", nil)
	rec := httptest.NewRecorder()

	handler.Credit(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestTransactionHandlerDebit_Success(t *testing.T) {
	handler := NewTransactionHandler(&fakeTransactionService{
		debitFn: func(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:         2,
				FromUserID: userID,
				Amount:     amount,
				Type:       domain.TransactionTypeDebit,
				Status:     domain.TransactionStatusCompleted,
				CreatedAt:  time.Now(),
			}, nil
		},
	})

	body := creditTransactionRequest{
		UserID: 1,
		Amount: 40,
	}

	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/debit", bytes.NewBuffer(payload))
	rec := httptest.NewRecorder()

	handler.Debit(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestTransactionHandlerDebit_InvalidMethod(t *testing.T) {
	handler := NewTransactionHandler(&fakeTransactionService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/debit", nil)
	rec := httptest.NewRecorder()

	handler.Debit(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("expected status %d, got %d", http.StatusMethodNotAllowed, rec.Code)
	}
}

func TestTransactionHandlerTransfer_Success(t *testing.T) {
	handler := NewTransactionHandler(&fakeTransactionService{
		transferFn: func(ctx context.Context, fromUserID, toUserID int64, amount float64) (*domain.Transaction, error) {
			return &domain.Transaction{
				ID:         3,
				FromUserID: fromUserID,
				ToUserID:   toUserID,
				Amount:     amount,
				Type:       domain.TransactionTypeTransfer,
				Status:     domain.TransactionStatusCompleted,
				CreatedAt:  time.Now(),
			}, nil
		},
	})

	body := transferTransactionRequest{
		FromUserID: 1,
		ToUserID:   2,
		Amount:     30,
	}

	payload, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/transfer", bytes.NewBuffer(payload))
	rec := httptest.NewRecorder()

	handler.Transfer(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}
}

func TestTransactionHandlerGetByID_Success(t *testing.T) {
	handler := NewTransactionHandler(&fakeTransactionService{
		transferFn: nil,
	})

	handler.transactionService = &fakeTransactionService{
		creditFn:   nil,
		debitFn:    nil,
		transferFn: nil,
	}

	handler.transactionService = &fakeTransactionService{
		creditFn:   nil,
		debitFn:    nil,
		transferFn: nil,
	}

	handler = NewTransactionHandler(&fakeTransactionService{
		transferFn: func(ctx context.Context, fromUserID, toUserID int64, amount float64) (*domain.Transaction, error) {
			return nil, nil
		},
	})

	handler = NewTransactionHandler(&fakeTransactionService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/1", nil)
	rec := httptest.NewRecorder()

	handler.GetByID(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rec.Code)
	}
}

func TestTransactionHandlerGetByUserID_Success(t *testing.T) {
	handler := NewTransactionHandler(&fakeTransactionService{})

	handler.transactionService = &fakeTransactionService{}

	handler = NewTransactionHandler(&fakeTransactionService{
		creditFn:   nil,
		debitFn:    nil,
		transferFn: nil,
	})

	handler = NewTransactionHandler(&fakeTransactionService{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/transactions/history?user_id=1", nil)
	rec := httptest.NewRecorder()

	handler.GetByUserID(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
