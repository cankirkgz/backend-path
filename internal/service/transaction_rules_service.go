package service

import (
	"context"
	"time"

	"backend-path/internal/domain"
	"backend-path/internal/interfaces"
)

type TransactionRulesConfig struct {
	MinAmount              float64
	MaxSingleAmount        float64
	MaxDailyOutgoingAmount float64
	MaxDailyCount          int
}

type TransactionRulesService struct {
	txRepo interfaces.TransactionRepository
	config TransactionRulesConfig
}

func NewTransactionRulesService(txRepo interfaces.TransactionRepository, config TransactionRulesConfig) *TransactionRulesService {
	return &TransactionRulesService{
		txRepo: txRepo,
		config: config,
	}
}

func DefaultTransactionRulesConfig() TransactionRulesConfig {
	return TransactionRulesConfig{
		MinAmount:              1,
		MaxSingleAmount:        10000,
		MaxDailyOutgoingAmount: 50000,
		MaxDailyCount:          20,
	}
}

func (s *TransactionRulesService) Validate(ctx context.Context, tx *domain.Transaction) error {
	if tx.Amount < s.config.MinAmount {
		return domain.ErrTransactionBelowMinimumAmount
	}

	if tx.Amount > s.config.MaxSingleAmount {
		return domain.ErrTransactionAboveMaximumAmount
	}

	userID := outgoingUserID(tx)
	if userID == 0 {
		return nil
	}

	todayTransactions, err := s.transactionsForToday(ctx, userID)
	if err != nil {
		return err
	}

	if len(todayTransactions) >= s.config.MaxDailyCount {
		return domain.ErrDailyTransactionCountExceeded
	}

	totalOutgoingAmount := 0.0

	for _, item := range todayTransactions {
		if item.Status != domain.TransactionStatusCompleted {
			continue
		}

		if item.Type == domain.TransactionTypeDebit && item.FromUserID == userID {
			totalOutgoingAmount += item.Amount
		}

		if item.Type == domain.TransactionTypeTransfer && item.FromUserID == userID {
			totalOutgoingAmount += item.Amount
		}
	}

	if totalOutgoingAmount+tx.Amount > s.config.MaxDailyOutgoingAmount {
		return domain.ErrDailyTransactionLimitExceeded
	}

	return nil
}

func (s *TransactionRulesService) transactionsForToday(ctx context.Context, userID int64) ([]*domain.Transaction, error) {
	transactions, err := s.txRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	result := make([]*domain.Transaction, 0)

	for _, tx := range transactions {
		if tx.CreatedAt.After(startOfDay) || tx.CreatedAt.Equal(startOfDay) {
			result = append(result, tx)
		}
	}

	return result, nil
}

func outgoingUserID(tx *domain.Transaction) int64 {
	if tx.Type == domain.TransactionTypeDebit || tx.Type == domain.TransactionTypeTransfer {
		return tx.FromUserID
	}

	return 0
}
