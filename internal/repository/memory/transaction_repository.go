package memory

import (
	"context"
	"sync"

	"backend-path/internal/domain"
)

type TransactionRepository struct {
	mu     sync.RWMutex
	data   map[int64]*domain.Transaction
	nextID int64
}

func NewTransactionRepository() *TransactionRepository {
	return &TransactionRepository{
		data:   make(map[int64]*domain.Transaction),
		nextID: 1,
	}
}

func (r *TransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx.ID = r.nextID
	r.nextID++

	copied := *tx
	r.data[tx.ID] = &copied

	return nil
}

func (r *TransactionRepository) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tx, ok := r.data[id]
	if !ok {
		return nil, nil
	}

	copied := *tx
	return &copied, nil
}

func (r *TransactionRepository) ListByUserID(ctx context.Context, userID int64) ([]*domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	transactions := make([]*domain.Transaction, 0)

	for _, tx := range r.data {
		if tx.FromUserID == userID || tx.ToUserID == userID {
			copied := *tx
			transactions = append(transactions, &copied)
		}
	}

	return transactions, nil
}

func (r *TransactionRepository) UpdateStatus(ctx context.Context, id int64, status domain.TransactionStatus) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, ok := r.data[id]
	if !ok {
		return nil
	}

	tx.Status = status
	return nil
}
