package service

import (
	"context"
	"strconv"
	"time"

	"backend-path/internal/domain"
	"backend-path/internal/interfaces"
)

type TransactionService struct {
	txRepo         interfaces.TransactionRepository
	balanceService interfaces.BalanceService
	eventStore     interfaces.EventStore
	rulesService   *TransactionRulesService
}

func NewTransactionService(
	txRepo interfaces.TransactionRepository,
	balanceService interfaces.BalanceService,
	eventStore interfaces.EventStore,
) *TransactionService {
	return &TransactionService{
		txRepo:         txRepo,
		balanceService: balanceService,
		eventStore:     eventStore,
		rulesService:   NewTransactionRulesService(txRepo, DefaultTransactionRulesConfig()),
	}
}

func (s *TransactionService) Credit(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error) {
	return s.CreditWithCurrency(ctx, userID, amount, domain.CurrencyTRY)
}

func (s *TransactionService) CreditWithCurrency(ctx context.Context, userID int64, amount float64, currency domain.Currency) (*domain.Transaction, error) {
	tx := &domain.Transaction{
		ToUserID:  userID,
		Amount:    amount,
		Currency:  currency,
		Type:      domain.TransactionTypeCredit,
		Status:    domain.TransactionStatusPending,
		CreatedAt: time.Now(),
	}

	if err := tx.Validate(); err != nil {
		return nil, err
	}

	if err := s.validateTransactionRules(ctx, tx); err != nil {
		return nil, err
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}

	if _, err := s.balanceService.CreditWithCurrency(ctx, userID, amount, currency); err != nil {
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

	if err := s.recordTransactionCreatedEvent(ctx, tx); err != nil {
		return nil, err
	}

	if err := s.recordBalanceUpdatedEvent(ctx, userID, amount, currency, "credit"); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *TransactionService) Debit(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error) {
	return s.DebitWithCurrency(ctx, userID, amount, domain.CurrencyTRY)
}

func (s *TransactionService) DebitWithCurrency(ctx context.Context, userID int64, amount float64, currency domain.Currency) (*domain.Transaction, error) {
	tx := &domain.Transaction{
		FromUserID: userID,
		Amount:     amount,
		Currency:   currency,
		Type:       domain.TransactionTypeDebit,
		Status:     domain.TransactionStatusPending,
		CreatedAt:  time.Now(),
	}

	if err := tx.Validate(); err != nil {
		return nil, err
	}

	if err := s.validateTransactionRules(ctx, tx); err != nil {
		return nil, err
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}

	if _, err := s.balanceService.DebitWithCurrency(ctx, userID, amount, currency); err != nil {
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

	if err := s.recordTransactionCreatedEvent(ctx, tx); err != nil {
		return nil, err
	}

	if err := s.recordBalanceUpdatedEvent(ctx, userID, -amount, currency, "debit"); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *TransactionService) Transfer(ctx context.Context, fromUserID, toUserID int64, amount float64) (*domain.Transaction, error) {
	return s.TransferWithCurrency(ctx, fromUserID, toUserID, amount, domain.CurrencyTRY)
}

func (s *TransactionService) TransferWithCurrency(ctx context.Context, fromUserID, toUserID int64, amount float64, currency domain.Currency) (*domain.Transaction, error) {
	tx := &domain.Transaction{
		FromUserID: fromUserID,
		ToUserID:   toUserID,
		Amount:     amount,
		Currency:   currency,
		Type:       domain.TransactionTypeTransfer,
		Status:     domain.TransactionStatusPending,
		CreatedAt:  time.Now(),
	}

	if err := tx.Validate(); err != nil {
		return nil, err
	}

	if err := s.validateTransactionRules(ctx, tx); err != nil {
		return nil, err
	}

	if err := s.txRepo.Create(ctx, tx); err != nil {
		return nil, err
	}

	if _, err := s.balanceService.DebitWithCurrency(ctx, fromUserID, amount, currency); err != nil {
		_ = tx.MarkFailed()
		_ = s.txRepo.UpdateStatus(ctx, tx.ID, tx.Status)
		return nil, err
	}

	if _, err := s.balanceService.CreditWithCurrency(ctx, toUserID, amount, currency); err != nil {
		_, _ = s.balanceService.CreditWithCurrency(ctx, fromUserID, amount, currency)

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

	if err := s.recordTransactionCreatedEvent(ctx, tx); err != nil {
		return nil, err
	}

	if err := s.recordBalanceUpdatedEvent(ctx, fromUserID, -amount, currency, "transfer_debit"); err != nil {
		return nil, err
	}

	if err := s.recordBalanceUpdatedEvent(ctx, toUserID, amount, currency, "transfer_credit"); err != nil {
		return nil, err
	}

	return tx, nil
}

func (s *TransactionService) BatchCredit(ctx context.Context, items []domain.BatchCreditItem) (*domain.BatchTransactionResult, error) {
	result := &domain.BatchTransactionResult{
		TotalCount: int64(len(items)),
		Items:      make([]domain.BatchTransactionItemResult, 0, len(items)),
	}

	for index, item := range items {
		tx, err := s.Credit(ctx, item.UserID, item.Amount)
		if err != nil {
			result.FailureCount++
			result.Items = append(result.Items, domain.BatchTransactionItemResult{
				Index:         int64(index),
				Error:         err.Error(),
				WasSuccessful: false,
			})
			continue
		}

		result.SuccessCount++
		result.Items = append(result.Items, domain.BatchTransactionItemResult{
			Index:         int64(index),
			Transaction:   tx,
			WasSuccessful: true,
		})
	}

	return result, nil
}

func (s *TransactionService) BatchDebit(ctx context.Context, items []domain.BatchDebitItem) (*domain.BatchTransactionResult, error) {
	result := &domain.BatchTransactionResult{
		TotalCount: int64(len(items)),
		Items:      make([]domain.BatchTransactionItemResult, 0, len(items)),
	}

	for index, item := range items {
		tx, err := s.Debit(ctx, item.UserID, item.Amount)
		if err != nil {
			result.FailureCount++
			result.Items = append(result.Items, domain.BatchTransactionItemResult{
				Index:         int64(index),
				Error:         err.Error(),
				WasSuccessful: false,
			})
			continue
		}

		result.SuccessCount++
		result.Items = append(result.Items, domain.BatchTransactionItemResult{
			Index:         int64(index),
			Transaction:   tx,
			WasSuccessful: true,
		})
	}

	return result, nil
}

func (s *TransactionService) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	return s.txRepo.GetByID(ctx, id)
}

func (s *TransactionService) GetByUserID(ctx context.Context, userID int64) ([]*domain.Transaction, error) {
	return s.txRepo.ListByUserID(ctx, userID)
}

func (s *TransactionService) recordTransactionCreatedEvent(ctx context.Context, tx *domain.Transaction) error {
	if s.eventStore == nil {
		return nil
	}

	return s.eventStore.Append(ctx, &domain.Event{
		Type:     domain.EventTransactionCreated,
		EntityID: stringFromInt64(tx.ID),
		Data: map[string]interface{}{
			"transaction_id": tx.ID,
			"from_user_id":   tx.FromUserID,
			"to_user_id":     tx.ToUserID,
			"amount":         tx.Amount,
			"currency":       tx.Currency,
			"type":           tx.Type,
			"status":         tx.Status,
		},
	})
}

func (s *TransactionService) recordBalanceUpdatedEvent(ctx context.Context, userID int64, amount float64, currency domain.Currency, reason string) error {
	if s.eventStore == nil {
		return nil
	}

	return s.eventStore.Append(ctx, &domain.Event{
		Type:     domain.EventBalanceUpdated,
		EntityID: stringFromInt64(userID),
		Data: map[string]interface{}{
			"user_id":  userID,
			"amount":   amount,
			"currency": currency,
			"reason":   reason,
		},
	})
}

func stringFromInt64(value int64) string {
	return strconv.FormatInt(value, 10)
}

func (s *TransactionService) validateTransactionRules(ctx context.Context, tx *domain.Transaction) error {
	if s.rulesService == nil {
		return nil
	}

	return s.rulesService.Validate(ctx, tx)
}
