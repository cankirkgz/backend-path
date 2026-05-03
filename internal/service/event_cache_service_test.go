package service

import (
	"context"
	"testing"

	"backend-path/internal/domain"
	"backend-path/internal/repository/memory"
)

func TestEventCacheServiceSyncBalanceCache(t *testing.T) {
	eventStore := memory.NewEventStoreRepository()
	cacheRepo := newFakeCacheRepository()

	_ = eventStore.Append(context.Background(), &domain.Event{
		Type: domain.EventBalanceUpdated,
		Data: map[string]interface{}{
			"user_id": int64(1),
			"amount":  100.0,
		},
	})

	_ = eventStore.Append(context.Background(), &domain.Event{
		Type: domain.EventBalanceUpdated,
		Data: map[string]interface{}{
			"user_id": int64(1),
			"amount":  -40.0,
		},
	})

	service := NewEventCacheService(eventStore, cacheRepo)

	err := service.SyncBalanceCache(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cached := cacheRepo.data["balance:1"].(*domain.Balance)

	if cached.Amount != 60 {
		t.Fatalf("expected balance=60, got %f", cached.Amount)
	}
}
