package memory

import (
	"context"
	"sync"
	"time"

	"backend-path/internal/domain"
)

type ScheduledTransactionRepository struct {
	mu     sync.RWMutex
	data   map[int64]*domain.ScheduledTransaction
	nextID int64
}

func NewScheduledTransactionRepository() *ScheduledTransactionRepository {
	return &ScheduledTransactionRepository{
		data:   make(map[int64]*domain.ScheduledTransaction),
		nextID: 1,
	}
}

func (r *ScheduledTransactionRepository) Create(ctx context.Context, tx *domain.ScheduledTransaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx.ID = r.nextID
	r.nextID++

	copied := *tx
	r.data[tx.ID] = &copied

	return nil
}

func (r *ScheduledTransactionRepository) GetByID(ctx context.Context, id int64) (*domain.ScheduledTransaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tx, ok := r.data[id]
	if !ok {
		return nil, nil
	}

	copied := *tx
	return &copied, nil
}

func (r *ScheduledTransactionRepository) ListDue(ctx context.Context, now time.Time) ([]*domain.ScheduledTransaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.ScheduledTransaction, 0)

	for _, tx := range r.data {
		if tx.Status != domain.ScheduledTransactionStatusPending {
			continue
		}

		if tx.RunAt.After(now) {
			continue
		}

		copied := *tx
		result = append(result, &copied)
	}

	return result, nil
}

func (r *ScheduledTransactionRepository) UpdateStatus(ctx context.Context, id int64, status domain.ScheduledTransactionStatus, processedAt *time.Time) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, ok := r.data[id]
	if !ok {
		return nil
	}

	tx.Status = status
	tx.ProcessedAt = processedAt

	return nil
}
