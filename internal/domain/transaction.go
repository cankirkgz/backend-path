package domain

import "time"

type TransactionType string
type TransactionStatus string

const (
	TransactionTypeCredit   TransactionType = "credit"
	TransactionTypeDebit    TransactionType = "debit"
	TransactionTypeTransfer TransactionType = "transfer"
)

const (
	TransactionStatusPending    TransactionStatus = "pending"
	TransactionStatusCompleted  TransactionStatus = "completed"
	TransactionStatusFailed     TransactionStatus = "failed"
	TransactionStatusRolledBack TransactionStatus = "rolled_back"
)

type Transaction struct {
	ID         int64             `json:"id"`
	FromUserID int64             `json:"from_user_id"`
	ToUserID   int64             `json:"to_user_id"`
	Amount     float64           `json:"amount"`
	Currency   Currency          `json:"currency"`
	Type       TransactionType   `json:"type"`
	Status     TransactionStatus `json:"status"`
	CreatedAt  time.Time         `json:"created_at"`
}

func (t *Transaction) Validate() error {
	if t.Amount <= 0 {
		return ErrInvalidTransactionAmount
	}
	if !t.Currency.IsValid() {
		return ErrInvalidCurrency
	}

	if t.Type != TransactionTypeCredit &&
		t.Type != TransactionTypeDebit &&
		t.Type != TransactionTypeTransfer {
		return ErrInvalidTransactionType
	}

	if t.Status != TransactionStatusPending &&
		t.Status != TransactionStatusCompleted &&
		t.Status != TransactionStatusFailed &&
		t.Status != TransactionStatusRolledBack {
		return ErrInvalidTransactionStatus
	}

	if t.Type == TransactionTypeTransfer {
		if t.FromUserID <= 0 || t.ToUserID <= 0 {
			return ErrInvalidTransactionUsers
		}

		if t.FromUserID == t.ToUserID {
			return ErrSameSenderReceiver
		}
	}

	if t.Type == TransactionTypeCredit && t.ToUserID <= 0 {
		return ErrInvalidTransactionUsers
	}

	if t.Type == TransactionTypeDebit && t.FromUserID <= 0 {
		return ErrInvalidTransactionUsers
	}

	return nil
}

func (t *Transaction) CanTransitionTo(newStatus TransactionStatus) bool {
	switch t.Status {
	case TransactionStatusPending:
		return newStatus == TransactionStatusCompleted ||
			newStatus == TransactionStatusFailed
	case TransactionStatusCompleted:
		return newStatus == TransactionStatusRolledBack
	default:
		return false
	}
}

func (t *Transaction) MarkCompleted() error {
	if !t.CanTransitionTo(TransactionStatusCompleted) {
		return ErrInvalidTransactionStateTransition
	}

	t.Status = TransactionStatusCompleted
	return nil
}

func (t *Transaction) MarkFailed() error {
	if !t.CanTransitionTo(TransactionStatusFailed) {
		return ErrInvalidTransactionStateTransition
	}

	t.Status = TransactionStatusFailed
	return nil
}

func (t *Transaction) MarkRolledBack() error {
	if !t.CanTransitionTo(TransactionStatusRolledBack) {
		return ErrInvalidTransactionStateTransition
	}

	t.Status = TransactionStatusRolledBack
	return nil
}
