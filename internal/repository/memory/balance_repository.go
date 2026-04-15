package memory

import (
	"context"
	"sync"
	"time"

	"backend-path/internal/domain"
)

type BalanceRepository struct {
	mu   sync.RWMutex
	data map[int64]*domain.Balance
}

func NewBalanceRepository() *BalanceRepository {
	return &BalanceRepository{
		data: make(map[int64]*domain.Balance),
	}
}

func (r *BalanceRepository) Create(ctx context.Context, balance *domain.Balance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *balance
	copied.LastUpdatedAt = time.Now()
	r.data[balance.UserID] = &copied

	return nil
}

func (r *BalanceRepository) GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	balance, ok := r.data[userID]
	if !ok {
		return nil, nil
	}

	copied := *balance
	return &copied, nil
}

func (r *BalanceRepository) Update(ctx context.Context, balance *domain.Balance) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *balance
	copied.LastUpdatedAt = time.Now()
	r.data[balance.UserID] = &copied

	return nil
}
