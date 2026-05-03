package domain

import "time"

type ScheduledTransactionStatus string

const (
	ScheduledTransactionStatusPending   ScheduledTransactionStatus = "pending"
	ScheduledTransactionStatusProcessed ScheduledTransactionStatus = "processed"
	ScheduledTransactionStatusFailed    ScheduledTransactionStatus = "failed"
)

type ScheduledTransaction struct {
	ID          int64                      `json:"id"`
	FromUserID  int64                      `json:"from_user_id"`
	ToUserID    int64                      `json:"to_user_id"`
	Amount      float64                    `json:"amount"`
	Type        TransactionType            `json:"type"`
	Status      ScheduledTransactionStatus `json:"status"`
	RunAt       time.Time                  `json:"run_at"`
	CreatedAt   time.Time                  `json:"created_at"`
	ProcessedAt *time.Time                 `json:"processed_at,omitempty"`
}

func (s *ScheduledTransaction) Validate() error {
	if s.Amount <= 0 {
		return ErrInvalidTransactionAmount
	}

	if s.RunAt.IsZero() {
		return ErrInvalidScheduledTransactionRunAt
	}

	if s.Type != TransactionTypeCredit &&
		s.Type != TransactionTypeDebit &&
		s.Type != TransactionTypeTransfer {
		return ErrInvalidTransactionType
	}

	if s.Type == TransactionTypeCredit && s.ToUserID <= 0 {
		return ErrInvalidTransactionUsers
	}

	if s.Type == TransactionTypeDebit && s.FromUserID <= 0 {
		return ErrInvalidTransactionUsers
	}

	if s.Type == TransactionTypeTransfer {
		if s.FromUserID <= 0 || s.ToUserID <= 0 {
			return ErrInvalidTransactionUsers
		}

		if s.FromUserID == s.ToUserID {
			return ErrSameSenderReceiver
		}
	}

	return nil
}
