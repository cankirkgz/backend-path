package interfaces

import (
	"context"

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
	Debit(ctx context.Context, userID int64, amount float64) (*domain.Transaction, error)
	Transfer(ctx context.Context, fromUserID, toUserID int64, amount float64) (*domain.Transaction, error)
	GetByID(ctx context.Context, id int64) (*domain.Transaction, error)
	GetByUserID(ctx context.Context, userID int64) ([]*domain.Transaction, error)
}

type BalanceService interface {
	GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error)
	Credit(ctx context.Context, userID int64, amount float64) (*domain.Balance, error)
	Debit(ctx context.Context, userID int64, amount float64) (*domain.Balance, error)
	GetHistory(ctx context.Context, userID int64) ([]*domain.AuditLog, error)
	GetCurrentAmount(ctx context.Context, userID int64) (float64, error)
}
