package service

import (
	"context"
	"fmt"
	"time"

	"backend-path/internal/domain"
	"backend-path/internal/interfaces"
)

type EventReplayService struct {
	eventStore interfaces.EventStore
}

func NewEventReplayService(eventStore interfaces.EventStore) *EventReplayService {
	return &EventReplayService{
		eventStore: eventStore,
	}
}

func (s *EventReplayService) RebuildBalances(ctx context.Context) (map[int64]*domain.Balance, error) {
	events, err := s.eventStore.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	balances := make(map[int64]*domain.Balance)

	for _, event := range events {
		if event.Type != domain.EventBalanceUpdated {
			continue
		}

		userID, err := getInt64FromEventData(event.Data, "user_id")
		if err != nil {
			return nil, err
		}

		amount, err := getFloat64FromEventData(event.Data, "amount")
		if err != nil {
			return nil, err
		}

		balance, ok := balances[userID]
		if !ok {
			balance = &domain.Balance{
				UserID:        userID,
				Amount:        0,
				LastUpdatedAt: time.Now(),
			}
			balances[userID] = balance
		}

		balance.Amount += amount
		balance.LastUpdatedAt = event.CreatedAt
	}

	return balances, nil
}

func getInt64FromEventData(data map[string]interface{}, key string) (int64, error) {
	value, ok := data[key]
	if !ok {
		return 0, fmt.Errorf("missing %s in event data", key)
	}

	switch v := value.(type) {
	case int64:
		return v, nil
	case int:
		return int64(v), nil
	case float64:
		return int64(v), nil
	default:
		return 0, fmt.Errorf("invalid %s type in event data", key)
	}
}

func getFloat64FromEventData(data map[string]interface{}, key string) (float64, error) {
	value, ok := data[key]
	if !ok {
		return 0, fmt.Errorf("missing %s in event data", key)
	}

	switch v := value.(type) {
	case float64:
		return v, nil
	case int:
		return float64(v), nil
	case int64:
		return float64(v), nil
	default:
		return 0, fmt.Errorf("invalid %s type in event data", key)
	}
}
