package interfaces

import (
	"context"
	"time"

	"backend-path/internal/domain"
)

type UserService interface {
	Register(ctx context.Context, user *domain.User, plainPassword string) error
	Authenticate(ctx context.Context, email, plainPassword string) (*domain.User, error)
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context) ([]*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int64) error
	AuthorizeAdmin(user *domain.User) error
}

type TransactionService interface {
	Credit(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error)
	CreditWithCurrency(ctx context.Context, userID int64, amount float64, currency domain.Currency) (*domain.Transaction, error)

	Debit(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error)
	DebitWithCurrency(ctx context.Context, userID int64, amount float64, currency domain.Currency) (*domain.Transaction, error)

	Transfer(ctx context.Context, fromUserID, toUserID int64, amount float64) (*domain.Transaction, error)
	TransferWithCurrency(ctx context.Context, fromUserID, toUserID int64, amount float64, currency domain.Currency) (*domain.Transaction, error)

	BatchCredit(ctx context.Context, items []domain.BatchCreditItem) (*domain.BatchTransactionResult, error)
	BatchDebit(ctx context.Context, items []domain.BatchDebitItem) (*domain.BatchTransactionResult, error)
	GetByID(ctx context.Context, id int64) (*domain.Transaction, error)
	GetByUserID(ctx context.Context, userID int64) ([]*domain.Transaction, error)
}

type BalanceService interface {
	GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error)
	GetByUserIDAndCurrency(ctx context.Context, userID int64, currency domain.Currency) (*domain.Balance, error)

	Credit(ctx context.Context, userID int64, amount float64) (*domain.Balance, error)
	CreditWithCurrency(ctx context.Context, userID int64, amount float64, currency domain.Currency) (*domain.Balance, error)

	Debit(ctx context.Context, userID int64, amount float64) (*domain.Balance, error)
	DebitWithCurrency(ctx context.Context, userID int64, amount float64, currency domain.Currency) (*domain.Balance, error)

	GetHistory(ctx context.Context, userID int64) ([]*domain.AuditLog, error)
	GetCurrentAmount(ctx context.Context, userID int64) (float64, error)
	GetCurrentAmountByCurrency(ctx context.Context, userID int64, currency domain.Currency) (float64, error)
}

type ScheduledTransactionService interface {
	Schedule(ctx context.Context, tx *domain.ScheduledTransaction) error
	ProcessDue(ctx context.Context, now time.Time) error
}
