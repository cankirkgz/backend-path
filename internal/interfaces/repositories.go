package interfaces

import (
	"context"

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
	Update(ctx context.Context, balance *domain.Balance) error
}

type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	ListByEntityID(ctx context.Context, entityType, entityID string) ([]*domain.AuditLog, error)
}
