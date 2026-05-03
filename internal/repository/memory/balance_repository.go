package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"backend-path/internal/domain"
)

type BalanceRepository struct {
	mu   sync.RWMutex
	data map[string]*domain.Balance
}

func NewBalanceRepository() *BalanceRepository {
	return &BalanceRepository{
		data: make(map[string]*domain.Balance),
	}
}

func (r *BalanceRepository) Create(ctx context.Context, balance *domain.Balance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if balance.Currency == "" {
		balance.Currency = domain.CurrencyTRY
	}

	copied := *balance
	copied.LastUpdatedAt = time.Now()

	r.data[balanceKey(balance.UserID, balance.Currency)] = &copied

	return nil
}

func (r *BalanceRepository) GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error) {
	return r.GetByUserIDAndCurrency(ctx, userID, domain.CurrencyTRY)
}

func (r *BalanceRepository) GetByUserIDAndCurrency(ctx context.Context, userID int64, currency domain.Currency) (*domain.Balance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if currency == "" {
		currency = domain.CurrencyTRY
	}

	balance, ok := r.data[balanceKey(userID, currency)]
	if !ok {
		return nil, nil
	}

	copied := *balance
	return &copied, nil
}

func (r *BalanceRepository) Update(ctx context.Context, balance *domain.Balance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if balance.Currency == "" {
		balance.Currency = domain.CurrencyTRY
	}

	copied := *balance
	copied.LastUpdatedAt = time.Now()

	r.data[balanceKey(balance.UserID, balance.Currency)] = &copied

	return nil
}

func balanceKey(userID int64, currency domain.Currency) string {
	return fmt.Sprintf("%d:%s", userID, currency)
}
