package server

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-path/internal/domain"
)

type fakeBalanceService struct {
	getCurrentAmountFn func(ctx context.Context, userID int64) (float64, error)
	getHistoryFn       func(ctx context.Context, userID int64) ([]*domain.AuditLog, error)
}

func (f *fakeBalanceService) GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error) {
	return nil, nil
}

func (f *fakeBalanceService) Credit(ctx context.Context, userID int64, amount float64) (*domain.Balance, error) {
	return nil, nil
}

func (f *fakeBalanceService) Debit(ctx context.Context, userID int64, amount float64) (*domain.Balance, error) {
	return nil, nil
}

func (f *fakeBalanceService) GetHistory(ctx context.Context, userID int64) ([]*domain.AuditLog, error) {
	if f.getHistoryFn == nil {
		return []*domain.AuditLog{}, nil
	}
	return f.getHistoryFn(ctx, userID)
}

func (f *fakeBalanceService) GetCurrentAmount(ctx context.Context, userID int64) (float64, error) {
	if f.getCurrentAmountFn == nil {
		return 0, nil
	}
	return f.getCurrentAmountFn(ctx, userID)
}

func TestBalanceHandlerGetCurrent_Success(t *testing.T) {
	handler := NewBalanceHandler(&fakeBalanceService{
		getCurrentAmountFn: func(ctx context.Context, userID int64) (float64, error) {
			return 70, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/balances/current?user_id=1", nil)
	rec := httptest.NewRecorder()

	handler.GetCurrent(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestBalanceHandlerGetHistorical_Success(t *testing.T) {
	handler := NewBalanceHandler(&fakeBalanceService{
		getHistoryFn: func(ctx context.Context, userID int64) ([]*domain.AuditLog, error) {
			return []*domain.AuditLog{
				{
					Action:    "credit",
					Details:   "credited 100.00, new_balance 100.00",
					CreatedAt: time.Now(),
				},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/balances/historical?user_id=1", nil)
	rec := httptest.NewRecorder()

	handler.GetHistorical(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestBalanceHandlerGetAtTime_Success(t *testing.T) {
	handler := NewBalanceHandler(&fakeBalanceService{
		getHistoryFn: func(ctx context.Context, userID int64) ([]*domain.AuditLog, error) {
			return []*domain.AuditLog{
				{
					Action:    "credit",
					Details:   "credited 100.00, new_balance 100.00",
					CreatedAt: time.Date(2026, 4, 14, 0, 0, 0, 0, time.UTC),
				},
			}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/balances/at-time?user_id=1&at=2026-04-14T00:10:00Z", nil)
	rec := httptest.NewRecorder()

	handler.GetAtTime(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func (f *fakeBalanceService) CreditWithCurrency(ctx context.Context, userID int64, amount float64, currency domain.Currency) (*domain.Balance, error) {
	return f.Credit(ctx, userID, amount)
}

func (f *fakeBalanceService) DebitWithCurrency(ctx context.Context, userID int64, amount float64, currency domain.Currency) (*domain.Balance, error) {
	return f.Debit(ctx, userID, amount)
}

func (f *fakeBalanceService) GetByUserIDAndCurrency(ctx context.Context, userID int64, currency domain.Currency) (*domain.Balance, error) {
	return f.GetByUserID(ctx, userID)
}

func (f *fakeBalanceService) GetCurrentAmountByCurrency(ctx context.Context, userID int64, currency domain.Currency) (float64, error) {
	return f.GetCurrentAmount(ctx, userID)
}
