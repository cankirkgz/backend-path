package interfaces

import (
	"context"
	"time"

	"backend-path/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id int64) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByUsername(ctx context.Context, username string) (*domain.User, error)
	List(ctx context.Context) ([]*domain.User, error)
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, id int64) error
}

type TransactionRepository interface {
	Create(ctx context.Context, tx *domain.Transaction) error
	GetByID(ctx context.Context, id int64) (*domain.Transaction, error)
	ListByUserID(ctx context.Context, userID int64) ([]*domain.Transaction, error)
	UpdateStatus(ctx context.Context, id int64, status domain.TransactionStatus) error
}

type BalanceRepository interface {
	Create(ctx context.Context, balance *domain.Balance) error
	GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error)
	GetByUserIDAndCurrency(ctx context.Context, userID int64, currency domain.Currency) (*domain.Balance, error)
	Update(ctx context.Context, balance *domain.Balance) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	ListByEntityID(ctx context.Context, entityType, entityID string) ([]*domain.AuditLog, error)
}

type EventStore interface {
	Append(ctx context.Context, event *domain.Event) error
	ListByEntityID(ctx context.Context, entityID string) ([]*domain.Event, error)
	ListAll(ctx context.Context) ([]*domain.Event, error)
}

type ScheduledTransactionRepository interface {
	Create(ctx context.Context, tx *domain.ScheduledTransaction) error
	GetByID(ctx context.Context, id int64) (*domain.ScheduledTransaction, error)
	ListDue(ctx context.Context, now time.Time) ([]*domain.ScheduledTransaction, error)
	UpdateStatus(ctx context.Context, id int64, status domain.ScheduledTransactionStatus, processedAt *time.Time) error
}
