package service

import (
	"context"
	"testing"
	"time"

	"backend-path/internal/domain"
)

func TestTransactionRulesService_AllowsValidDebit(t *testing.T) {
	txRepo := newFakeTransactionRepository()

	rulesService := NewTransactionRulesService(txRepo, TransactionRulesConfig{
		MinAmount:              1,
		MaxSingleAmount:        1000,
		MaxDailyOutgoingAmount: 5000,
		MaxDailyCount:          5,
	})

	tx := &domain.Transaction{
		FromUserID: 1,
		Amount:     100,
		Type:       domain.TransactionTypeDebit,
		Status:     domain.TransactionStatusPending,
		CreatedAt:  time.Now(),
	}

	err := rulesService.Validate(context.Background(), tx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestTransactionRulesService_RejectsBelowMinimumAmount(t *testing.T) {
	txRepo := newFakeTransactionRepository()

	rulesService := NewTransactionRulesService(txRepo, TransactionRulesConfig{
		MinAmount:              10,
		MaxSingleAmount:        1000,
		MaxDailyOutgoingAmount: 5000,
		MaxDailyCount:          5,
	})

	tx := &domain.Transaction{
		FromUserID: 1,
		Amount:     5,
		Type:       domain.TransactionTypeDebit,
		Status:     domain.TransactionStatusPending,
		CreatedAt:  time.Now(),
	}

	err := rulesService.Validate(context.Background(), tx)
	if err != domain.ErrTransactionBelowMinimumAmount {
		t.Fatalf("expected ErrTransactionBelowMinimumAmount, got %v", err)
	}
}

func TestTransactionRulesService_RejectsAboveMaximumAmount(t *testing.T) {
	txRepo := newFakeTransactionRepository()

	rulesService := NewTransactionRulesService(txRepo, TransactionRulesConfig{
		MinAmount:              1,
		MaxSingleAmount:        1000,
		MaxDailyOutgoingAmount: 5000,
		MaxDailyCount:          5,
	})

	tx := &domain.Transaction{
		FromUserID: 1,
		Amount:     1500,
		Type:       domain.TransactionTypeDebit,
		Status:     domain.TransactionStatusPending,
		CreatedAt:  time.Now(),
	}

	err := rulesService.Validate(context.Background(), tx)
	if err != domain.ErrTransactionAboveMaximumAmount {
		t.Fatalf("expected ErrTransactionAboveMaximumAmount, got %v", err)
	}
}

func TestTransactionRulesService_RejectsDailyOutgoingAmountLimit(t *testing.T) {
	txRepo := newFakeTransactionRepository()

	existingTx := &domain.Transaction{
		FromUserID: 1,
		Amount:     900,
		Type:       domain.TransactionTypeDebit,
		Status:     domain.TransactionStatusCompleted,
		CreatedAt:  time.Now(),
	}

	if err := txRepo.Create(context.Background(), existingTx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	rulesService := NewTransactionRulesService(txRepo, TransactionRulesConfig{
		MinAmount:              1,
		MaxSingleAmount:        1000,
		MaxDailyOutgoingAmount: 1000,
		MaxDailyCount:          5,
	})

	newTx := &domain.Transaction{
		FromUserID: 1,
		Amount:     200,
		Type:       domain.TransactionTypeDebit,
		Status:     domain.TransactionStatusPending,
		CreatedAt:  time.Now(),
	}

	err := rulesService.Validate(context.Background(), newTx)
	if err != domain.ErrDailyTransactionLimitExceeded {
		t.Fatalf("expected ErrDailyTransactionLimitExceeded, got %v", err)
	}
}

func TestTransactionRulesService_RejectsDailyTransactionCountLimit(t *testing.T) {
	txRepo := newFakeTransactionRepository()

	for i := 0; i < 2; i++ {
		existingTx := &domain.Transaction{
			FromUserID: 1,
			Amount:     100,
			Type:       domain.TransactionTypeDebit,
			Status:     domain.TransactionStatusCompleted,
			CreatedAt:  time.Now(),
		}

		if err := txRepo.Create(context.Background(), existingTx); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

	rulesService := NewTransactionRulesService(txRepo, TransactionRulesConfig{
		MinAmount:              1,
		MaxSingleAmount:        1000,
		MaxDailyOutgoingAmount: 5000,
		MaxDailyCount:          2,
	})

	newTx := &domain.Transaction{
		FromUserID: 1,
		Amount:     100,
		Type:       domain.TransactionTypeDebit,
		Status:     domain.TransactionStatusPending,
		CreatedAt:  time.Now(),
	}

	err := rulesService.Validate(context.Background(), newTx)
	if err != domain.ErrDailyTransactionCountExceeded {
		t.Fatalf("expected ErrDailyTransactionCountExceeded, got %v", err)
	}
}

func TestTransactionRulesService_IgnoresOldTransactions(t *testing.T) {
	txRepo := newFakeTransactionRepository()

	oldTx := &domain.Transaction{
		FromUserID: 1,
		Amount:     900,
		Type:       domain.TransactionTypeDebit,
		Status:     domain.TransactionStatusCompleted,
		CreatedAt:  time.Now().Add(-48 * time.Hour),
	}

	if err := txRepo.Create(context.Background(), oldTx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	rulesService := NewTransactionRulesService(txRepo, TransactionRulesConfig{
		MinAmount:              1,
		MaxSingleAmount:        1000,
		MaxDailyOutgoingAmount: 1000,
		MaxDailyCount:          1,
	})

	newTx := &domain.Transaction{
		FromUserID: 1,
		Amount:     200,
		Type:       domain.TransactionTypeDebit,
		Status:     domain.TransactionStatusPending,
		CreatedAt:  time.Now(),
	}

	err := rulesService.Validate(context.Background(), newTx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestTransactionRulesService_AllowsCreditWithoutOutgoingLimitCheck(t *testing.T) {
	txRepo := newFakeTransactionRepository()

	rulesService := NewTransactionRulesService(txRepo, TransactionRulesConfig{
		MinAmount:              1,
		MaxSingleAmount:        1000,
		MaxDailyOutgoingAmount: 100,
		MaxDailyCount:          1,
	})

	tx := &domain.Transaction{
		ToUserID:  1,
		Amount:    900,
		Type:      domain.TransactionTypeCredit,
		Status:    domain.TransactionStatusPending,
		CreatedAt: time.Now(),
	}

	err := rulesService.Validate(context.Background(), tx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
