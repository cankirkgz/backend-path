package service

import (
	"context"
	"time"

	"backend-path/internal/domain"
	"backend-path/internal/interfaces"
)

type EventCacheService struct {
	eventStore interfaces.EventStore
	cacheRepo  interfaces.CacheRepository
}

func NewEventCacheService(
	eventStore interfaces.EventStore,
	cacheRepo interfaces.CacheRepository,
) *EventCacheService {
	return &EventCacheService{
		eventStore: eventStore,
		cacheRepo:  cacheRepo,
	}
}

const eventCacheTTL = 5 * time.Minute

func (s *EventCacheService) SyncBalanceCache(ctx context.Context) error {
	events, err := s.eventStore.ListAll(ctx)
	if err != nil {
		return err
	}

	balances := make(map[int64]float64)

	for _, event := range events {
		if event.Type != domain.EventBalanceUpdated {
			continue
		}

		userID, ok := event.Data["user_id"].(int64)
		if !ok {
			continue
		}

		amount, ok := event.Data["amount"].(float64)
		if !ok {
			continue
		}

		balances[userID] += amount
	}

	for userID, amount := range balances {
		balance := &domain.Balance{
			UserID: userID,
			Amount: amount,
		}

		_ = s.cacheRepo.Set(ctx, balanceCacheKey(userID), balance, eventCacheTTL)
	}

	return nil
}
