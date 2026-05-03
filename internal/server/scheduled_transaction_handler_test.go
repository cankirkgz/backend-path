package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"backend-path/internal/domain"
)

type fakeScheduledTransactionService struct {
	scheduledTx *domain.ScheduledTransaction
	processErr  error
}

func (s *fakeScheduledTransactionService) Schedule(ctx context.Context, tx *domain.ScheduledTransaction) error {
	if err := tx.Validate(); err != nil {
		return err
	}

	tx.ID = 1
	tx.Status = domain.ScheduledTransactionStatusPending
	tx.CreatedAt = time.Now()

	s.scheduledTx = tx
	return nil
}

func (s *fakeScheduledTransactionService) ProcessDue(ctx context.Context, now time.Time) error {
	return s.processErr
}

func TestScheduledTransactionHandlerSchedule_Success(t *testing.T) {
	service := &fakeScheduledTransactionService{}
	handler := NewScheduledTransactionHandler(service)

	body := strings.NewReader(`{
		"to_user_id": 1,
		"amount": 100,
		"type": "credit",
		"run_at": "2030-01-01T10:00:00Z"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scheduled-transactions", body)
	rec := httptest.NewRecorder()

	handler.Schedule(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, rec.Code)
	}

	var response domain.ScheduledTransaction
	if err := json.NewDecoder(rec.Body).Decode(&response); err != nil {
		t.Fatalf("expected valid response body, got %v", err)
	}

	if response.ID != 1 {
		t.Fatalf("expected id 1, got %d", response.ID)
	}

	if response.Status != domain.ScheduledTransactionStatusPending {
		t.Fatalf("expected pending status, got %s", response.Status)
	}

	if response.ToUserID != 1 {
		t.Fatalf("expected to_user_id 1, got %d", response.ToUserID)
	}

	if response.Amount != 100 {
		t.Fatalf("expected amount 100, got %f", response.Amount)
	}
}

func TestScheduledTransactionHandlerSchedule_RejectsInvalidJSON(t *testing.T) {
	handler := NewScheduledTransactionHandler(&fakeScheduledTransactionService{})

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/v1/scheduled-transactions",
		strings.NewReader(`invalid-json`),
	)
	rec := httptest.NewRecorder()

	handler.Schedule(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestScheduledTransactionHandlerSchedule_RejectsInvalidRunAt(t *testing.T) {
	handler := NewScheduledTransactionHandler(&fakeScheduledTransactionService{})

	body := strings.NewReader(`{
		"to_user_id": 1,
		"amount": 100,
		"type": "credit",
		"run_at": "wrong-date"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scheduled-transactions", body)
	rec := httptest.NewRecorder()

	handler.Schedule(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestScheduledTransactionHandlerSchedule_RejectsInvalidAmount(t *testing.T) {
	handler := NewScheduledTransactionHandler(&fakeScheduledTransactionService{})

	body := strings.NewReader(`{
		"to_user_id": 1,
		"amount": 0,
		"type": "credit",
		"run_at": "2030-01-01T10:00:00Z"
	}`)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scheduled-transactions", body)
	rec := httptest.NewRecorder()

	handler.Schedule(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestScheduledTransactionHandlerProcessDue_Success(t *testing.T) {
	handler := NewScheduledTransactionHandler(&fakeScheduledTransactionService{})

	req := httptest.NewRequest(http.MethodPost, "/api/v1/scheduled-transactions/process-due", nil)
	rec := httptest.NewRecorder()

	handler.ProcessDue(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
}
