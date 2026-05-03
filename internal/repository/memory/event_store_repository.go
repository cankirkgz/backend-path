package memory

import (
	"context"
	"sync"
	"time"

	"backend-path/internal/domain"
)

type EventStoreRepository struct {
	mu     sync.RWMutex
	events []*domain.Event
	nextID int64
}

func NewEventStoreRepository() *EventStoreRepository {
	return &EventStoreRepository{
		events: make([]*domain.Event, 0),
		nextID: 1,
	}
}

func (r *EventStoreRepository) Append(ctx context.Context, event *domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	copied := *event
	copied.ID = r.nextID
	copied.CreatedAt = time.Now()

	r.nextID++
	r.events = append(r.events, &copied)

	return nil
}

func (r *EventStoreRepository) ListByEntityID(ctx context.Context, entityID string) ([]*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.Event, 0)

	for _, event := range r.events {
		if event.EntityID == entityID {
			copied := *event
			result = append(result, &copied)
		}
	}

	return result, nil
}

func (r *EventStoreRepository) ListAll(ctx context.Context) ([]*domain.Event, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*domain.Event, 0, len(r.events))

	for _, event := range r.events {
		copied := *event
		result = append(result, &copied)
	}

	return result, nil
}
