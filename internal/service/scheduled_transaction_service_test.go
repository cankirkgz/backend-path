package service

import (
	"context"
	"testing"
	"time"

	"backend-path/internal/domain"
	"backend-path/internal/repository/memory"
)

func TestScheduledTransactionServiceSchedule_Success(t *testing.T) {
	scheduledRepo := memory.NewScheduledTransactionRepository()
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()
	transactionService := NewTransactionService(txRepo, balanceService, nil)

	service := NewScheduledTransactionService(scheduledRepo, transactionService)

	scheduledTx := &domain.ScheduledTransaction{
		ToUserID: 1,
		Amount:   100,
		Type:     domain.TransactionTypeCredit,
		RunAt:    time.Now().Add(1 * time.Hour),
	}

	err := service.Schedule(context.Background(), scheduledTx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	storedTx, err := scheduledRepo.GetByID(context.Background(), scheduledTx.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if storedTx == nil {
		t.Fatal("expected scheduled transaction to be stored")
	}

	if storedTx.Status != domain.ScheduledTransactionStatusPending {
		t.Fatalf("expected pending status, got %s", storedTx.Status)
	}
}

func TestScheduledTransactionServiceSchedule_InvalidAmount(t *testing.T) {
	scheduledRepo := memory.NewScheduledTransactionRepository()
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()
	transactionService := NewTransactionService(txRepo, balanceService, nil)

	service := NewScheduledTransactionService(scheduledRepo, transactionService)

	scheduledTx := &domain.ScheduledTransaction{
		ToUserID: 1,
		Amount:   0,
		Type:     domain.TransactionTypeCredit,
		RunAt:    time.Now().Add(1 * time.Hour),
	}

	err := service.Schedule(context.Background(), scheduledTx)
	if err != domain.ErrInvalidTransactionAmount {
		t.Fatalf("expected ErrInvalidTransactionAmount, got %v", err)
	}
}

func TestScheduledTransactionServiceProcessDue_ProcessesCredit(t *testing.T) {
	scheduledRepo := memory.NewScheduledTransactionRepository()
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()
	transactionService := NewTransactionService(txRepo, balanceService, nil)

	service := NewScheduledTransactionService(scheduledRepo, transactionService)

	scheduledTx := &domain.ScheduledTransaction{
		ToUserID: 1,
		Amount:   100,
		Type:     domain.TransactionTypeCredit,
		RunAt:    time.Now().Add(-1 * time.Minute),
	}

	if err := service.Schedule(context.Background(), scheduledTx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err := service.ProcessDue(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	storedTx, err := scheduledRepo.GetByID(context.Background(), scheduledTx.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if storedTx.Status != domain.ScheduledTransactionStatusProcessed {
		t.Fatalf("expected processed status, got %s", storedTx.Status)
	}

	if storedTx.ProcessedAt == nil {
		t.Fatal("expected processed_at to be set")
	}

	if balanceService.balances[1].Amount != 100 {
		t.Fatalf("expected balance 100, got %f", balanceService.balances[1].Amount)
	}
}

func TestScheduledTransactionServiceProcessDue_DoesNotProcessFutureTransaction(t *testing.T) {
	scheduledRepo := memory.NewScheduledTransactionRepository()
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()
	transactionService := NewTransactionService(txRepo, balanceService, nil)

	service := NewScheduledTransactionService(scheduledRepo, transactionService)

	scheduledTx := &domain.ScheduledTransaction{
		ToUserID: 1,
		Amount:   100,
		Type:     domain.TransactionTypeCredit,
		RunAt:    time.Now().Add(1 * time.Hour),
	}

	if err := service.Schedule(context.Background(), scheduledTx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err := service.ProcessDue(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	storedTx, err := scheduledRepo.GetByID(context.Background(), scheduledTx.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if storedTx.Status != domain.ScheduledTransactionStatusPending {
		t.Fatalf("expected pending status, got %s", storedTx.Status)
	}

	if len(balanceService.balances) != 0 {
		t.Fatalf("expected no balance update, got %d", len(balanceService.balances))
	}
}

func TestScheduledTransactionServiceProcessDue_MarksFailedWhenTransactionFails(t *testing.T) {
	scheduledRepo := memory.NewScheduledTransactionRepository()
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()
	transactionService := NewTransactionService(txRepo, balanceService, nil)

	service := NewScheduledTransactionService(scheduledRepo, transactionService)

	scheduledTx := &domain.ScheduledTransaction{
		FromUserID: 1,
		Amount:     100,
		Type:       domain.TransactionTypeDebit,
		RunAt:      time.Now().Add(-1 * time.Minute),
	}

	if err := service.Schedule(context.Background(), scheduledTx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err := service.ProcessDue(context.Background(), time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	storedTx, err := scheduledRepo.GetByID(context.Background(), scheduledTx.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if storedTx.Status != domain.ScheduledTransactionStatusFailed {
		t.Fatalf("expected failed status, got %s", storedTx.Status)
	}

	if storedTx.ProcessedAt == nil {
		t.Fatal("expected processed_at to be set")
	}
}
