package memory

import (
	"context"
	"sync"

	"backend-path/internal/domain"
)

type AuditLogRepository struct {
	mu     sync.RWMutex
	data   []*domain.AuditLog
	nextID int64
}

func NewAuditLogRepository() *AuditLogRepository {
	return &AuditLogRepository{
		data:   make([]*domain.AuditLog, 0),
		nextID: 1,
	}
}

func (r *AuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *log
	copied.ID = r.nextID
	r.nextID++

	r.data = append(r.data, &copied)
	return nil
}

func (r *AuditLogRepository) ListByEntityID(ctx context.Context, entityType, entityID string) ([]*domain.AuditLog, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.AuditLog, 0)

	for _, log := range r.data {
		if log.EntityType == entityType && log.EntityID == entityID {
			copied := *log
			result = append(result, &copied)
		}
	}

	return result, nil
}
