package service

import (
	"context"
	"time"

	"backend-path/internal/domain"
	"backend-path/internal/interfaces"
)

type ScheduledTransactionService struct {
	scheduledRepo      interfaces.ScheduledTransactionRepository
	transactionService interfaces.TransactionService
}

func NewScheduledTransactionService(
	scheduledRepo interfaces.ScheduledTransactionRepository,
	transactionService interfaces.TransactionService,
) *ScheduledTransactionService {
	return &ScheduledTransactionService{
		scheduledRepo:      scheduledRepo,
		transactionService: transactionService,
	}
}

func (s *ScheduledTransactionService) Schedule(ctx context.Context, tx *domain.ScheduledTransaction) error {
	if tx.CreatedAt.IsZero() {
		tx.CreatedAt = time.Now()
	}

	tx.Status = domain.ScheduledTransactionStatusPending

	if err := tx.Validate(); err != nil {
		return err
	}

	return s.scheduledRepo.Create(ctx, tx)
}

func (s *ScheduledTransactionService) ProcessDue(ctx context.Context, now time.Time) error {
	dueTransactions, err := s.scheduledRepo.ListDue(ctx, now)
	if err != nil {
		return err
	}

	for _, scheduledTx := range dueTransactions {
		processedAt := time.Now()

		if err := s.processOne(ctx, scheduledTx); err != nil {
			_ = s.scheduledRepo.UpdateStatus(
				ctx,
				scheduledTx.ID,
				domain.ScheduledTransactionStatusFailed,
				&processedAt,
			)
			continue
		}

		_ = s.scheduledRepo.UpdateStatus(
			ctx,
			scheduledTx.ID,
			domain.ScheduledTransactionStatusProcessed,
			&processedAt,
		)
	}

	return nil
}

func (s *ScheduledTransactionService) processOne(ctx context.Context, scheduledTx *domain.ScheduledTransaction) error {
	switch scheduledTx.Type {
	case domain.TransactionTypeCredit:
		_, err := s.transactionService.Credit(ctx, scheduledTx.ToUserID, scheduledTx.Amount)
		return err

	case domain.TransactionTypeDebit:
		_, err := s.transactionService.Debit(ctx, scheduledTx.FromUserID, scheduledTx.Amount)
		return err

	case domain.TransactionTypeTransfer:
		_, err := s.transactionService.Transfer(
			ctx,
			scheduledTx.FromUserID,
			scheduledTx.ToUserID,
			scheduledTx.Amount,
		)
		return err

	default:
		return domain.ErrInvalidTransactionType
	}
}
