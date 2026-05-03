package server

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
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

func (f *fakeTransactionService) CreditWithCurrency(ctx context.Context, userID int64, amount float64, currency domain.Currency) (*domain.Transaction, error) {
	return f.Credit(ctx, userID, amount)
}

func (f *fakeTransactionService) DebitWithCurrency(ctx context.Context, userID int64, amount float64, currency domain.Currency) (*domain.Transaction, error) {
	return f.Debit(ctx, userID, amount)
}

func (f *fakeTransactionService) TransferWithCurrency(ctx context.Context, fromUserID, toUserID int64, amount float64, currency domain.Currency) (*domain.Transaction, error) {
	return f.Transfer(ctx, fromUserID, toUserID, amount)
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

func (s *fakeTransactionService) BatchCredit(ctx context.Context, items []domain.BatchCreditItem) (*domain.BatchTransactionResult, error) {
	result := &domain.BatchTransactionResult{
		TotalCount: int64(len(items)),
		Items:      make([]domain.BatchTransactionItemResult, 0, len(items)),
	}

	for index, item := range items {
		if item.Amount <= 0 {
			result.FailureCount++
			result.Items = append(result.Items, domain.BatchTransactionItemResult{
				Index:         int64(index),
				Error:         domain.ErrInvalidTransactionAmount.Error(),
				WasSuccessful: false,
			})
			continue
		}

		result.SuccessCount++
		result.Items = append(result.Items, domain.BatchTransactionItemResult{
			Index: int64(index),
			Transaction: &domain.Transaction{
				ID:       int64(index + 1),
				ToUserID: item.UserID,
				Amount:   item.Amount,
				Type:     domain.TransactionTypeCredit,
				Status:   domain.TransactionStatusCompleted,
			},
			WasSuccessful: true,
		})
	}

	return result, nil
}

func (s *fakeTransactionService) BatchDebit(ctx context.Context, items []domain.BatchDebitItem) (*domain.BatchTransactionResult, error) {
	result := &domain.BatchTransactionResult{
		TotalCount: int64(len(items)),
		Items:      make([]domain.BatchTransactionItemResult, 0, len(items)),
	}

	for index, item := range items {
		if item.Amount <= 0 {
			result.FailureCount++
			result.Items = append(result.Items, domain.BatchTransactionItemResult{
				Index:         int64(index),
				Error:         domain.ErrInvalidTransactionAmount.Error(),
				WasSuccessful: false,
			})
			continue
		}

		result.SuccessCount++
		result.Items = append(result.Items, domain.BatchTransactionItemResult{
			Index: int64(index),
			Transaction: &domain.Transaction{
				ID:         int64(index + 1),
				FromUserID: item.UserID,
				Amount:     item.Amount,
				Type:       domain.TransactionTypeDebit,
				Status:     domain.TransactionStatusCompleted,
			},
			WasSuccessful: true,
		})
	}

	return result, nil
}

func TestTransactionHandlerBatchCredit_Success(t *testing.T) {
	handler := NewTransactionHandler(&fakeTransactionService{})

	body := strings.NewReader(`{
		"items": [
			{"user_id": 1, "amount": 100},
			{"user_id": 2, "amount": 200},
			{"user_id": 3, "amount": 0}
		]
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/batch-credit", body)
	rec := httptest.NewRecorder()

	handler.BatchCredit(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response domain.BatchTransactionResult
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("expected valid response body, got %v", err)
	}

	if response.TotalCount != 3 {
		t.Fatalf("expected total count 3, got %d", response.TotalCount)
	}

	if response.SuccessCount != 2 {
		t.Fatalf("expected success count 2, got %d", response.SuccessCount)
	}

	if response.FailureCount != 1 {
		t.Fatalf("expected failure count 1, got %d", response.FailureCount)
	}
}

func TestTransactionHandlerBatchDebit_Success(t *testing.T) {
	handler := NewTransactionHandler(&fakeTransactionService{})

	body := strings.NewReader(`{
		"items": [
			{"user_id": 1, "amount": 100},
			{"user_id": 2, "amount": 0}
		]
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/batch-debit", body)
	rec := httptest.NewRecorder()

	handler.BatchDebit(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var response domain.BatchTransactionResult
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("expected valid response body, got %v", err)
	}

	if response.TotalCount != 2 {
		t.Fatalf("expected total count 2, got %d", response.TotalCount)
	}

	if response.SuccessCount != 1 {
		t.Fatalf("expected success count 1, got %d", response.SuccessCount)
	}

	if response.FailureCount != 1 {
		t.Fatalf("expected failure count 1, got %d", response.FailureCount)
	}
}

func TestTransactionHandlerBatchCredit_RejectsInvalidJSON(t *testing.T) {
	handler := NewTransactionHandler(&fakeTransactionService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/batch-credit", strings.NewReader(`invalid-json`))
	rec := httptest.NewRecorder()

	handler.BatchCredit(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestTransactionHandlerBatchDebit_RejectsInvalidJSON(t *testing.T) {
	handler := NewTransactionHandler(&fakeTransactionService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/transactions/batch-debit", strings.NewReader(`invalid-json`))
	rec := httptest.NewRecorder()

	handler.BatchDebit(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}
