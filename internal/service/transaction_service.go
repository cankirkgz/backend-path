package service

import (
	"context"
	"time"

	"backend-path/internal/domain"
	"backend-path/internal/interfaces"
)

type TransactionService struct {
	txRepo         interfaces.TransactionRepository
	balanceService interfaces.BalanceService
}

func NewTransactionService(
	txRepo interfaces.TransactionRepository,
	balanceService interfaces.BalanceService,
) *TransactionService {
	return &TransactionService{
		txRepo:         txRepo,
		balanceService: balanceService,
	}
}

func (s *TransactionService) Credit(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error) {
	tx := &domain.Transaction{
		ToUserID:  userID,
		Amount:    amount,
		Type:      domain.TransactionTypeCredit,
		Status:    domain.TransactionStatusPending,
		CreatedAt: time.Now(),
	}

	if err := tx.Validate(); err != nil {
		return nil, err
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}

	if _, err := s.balanceService.Credit(ctx, userID, amount); err != nil {
		_ = tx.MarkFailed()
		_ = s.txRepo.UpdateStatus(ctx, tx.ID, tx.Status)
		return nil, err
	}

	if err := tx.MarkCompleted(); err != nil {
		return nil, err
	}

	if err := s.txRepo.UpdateStatus(ctx, tx.ID, tx.Status); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *TransactionService) Debit(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error) {
	tx := &domain.Transaction{
		FromUserID: userID,
		Amount:     amount,
		Type:       domain.TransactionTypeDebit,
		Status:     domain.TransactionStatusPending,
		CreatedAt:  time.Now(),
	}

	if err := tx.Validate(); err != nil {
		return nil, err
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}

	if _, err := s.balanceService.Debit(ctx, userID, amount); err != nil {
		_ = tx.MarkFailed()
		_ = s.txRepo.UpdateStatus(ctx, tx.ID, tx.Status)
		return nil, err
	}

	if err := tx.MarkCompleted(); err != nil {
		return nil, err
	}

	if err := s.txRepo.UpdateStatus(ctx, tx.ID, tx.Status); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *TransactionService) Transfer(ctx context.Context, fromUserID, toUserID int64, amount float64) (*domain.Transaction, error) {
	tx := &domain.Transaction{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Amount:     amount,
		Type:       domain.TransactionTypeTransfer,
		Status:     domain.TransactionStatusPending,
		CreatedAt:  time.Now(),
	}

	if err := tx.Validate(); err != nil {
		return nil, err
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}

	if _, err := s.balanceService.Debit(ctx, fromUserID, amount); err != nil {
		_ = tx.MarkFailed()
		_ = s.txRepo.UpdateStatus(ctx, tx.ID, tx.Status)
		return nil, err
	}

	if _, err := s.balanceService.Credit(ctx, toUserID, amount); err != nil {
		// basic rollback: return money to sender
		_, _ = s.balanceService.Credit(ctx, fromUserID, amount)

		_ = tx.MarkFailed()
		_ = s.txRepo.UpdateStatus(ctx, tx.ID, tx.Status)
		return nil, err
	}

	if err := tx.MarkCompleted(); err != nil {
		return nil, err
	}

	if err := s.txRepo.UpdateStatus(ctx, tx.ID, tx.Status); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *TransactionService) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	return s.txRepo.GetByID(ctx, id)
}

func (s *TransactionService) GetByUserID(ctx context.Context, userID int64) ([]*domain.Transaction, error) {
	return s.txRepo.ListByUserID(ctx, userID)
}
