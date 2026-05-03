package service

import (
	"context"
	"testing"
	"time"

	"backend-path/internal/domain"
)

func TestEventReplayService_RebuildBalances(t *testing.T) {
	eventStore := newFakeEventStore()

	_ = eventStore.Append(context.Background(), &domain.Event{
		Type:      domain.EventBalanceUpdated,
		EntityID:  "1",
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"user_id": int64(1),
			"amount":  float64(100),
		},
	})

	_ = eventStore.Append(context.Background(), &domain.Event{
		Type:      domain.EventBalanceUpdated,
		EntityID:  "1",
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"user_id": int64(1),
			"amount":  float64(-40),
		},
	})

	_ = eventStore.Append(context.Background(), &domain.Event{
		Type:      domain.EventBalanceUpdated,
		EntityID:  "2",
		CreatedAt: time.Now(),
		Data: map[string]interface{}{
			"user_id": int64(2),
			"amount":  float64(75),
		},
	})

	replayService := NewEventReplayService(eventStore)

	balances, err := replayService.RebuildBalances(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if balances[1].Amount != 60 {
		t.Fatalf("expected user 1 balance 60, got %f", balances[1].Amount)
	}

	if balances[2].Amount != 75 {
		t.Fatalf("expected user 2 balance 75, got %f", balances[2].Amount)
	}
}

func TestEventReplayService_RebuildBalances_ReturnsErrorWhenUserIDMissing(t *testing.T) {
	eventStore := newFakeEventStore()

	_ = eventStore.Append(context.Background(), &domain.Event{
		Type: domain.EventBalanceUpdated,
		Data: map[string]interface{}{
			"amount": float64(100),
		},
	})

	replayService := NewEventReplayService(eventStore)

	_, err := replayService.RebuildBalances(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestEventReplayService_RebuildBalances_ReturnsErrorWhenAmountMissing(t *testing.T) {
	eventStore := newFakeEventStore()

	_ = eventStore.Append(context.Background(), &domain.Event{
		Type: domain.EventBalanceUpdated,
		Data: map[string]interface{}{
			"user_id": int64(1),
		},
	})

	replayService := NewEventReplayService(eventStore)

	_, err := replayService.RebuildBalances(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
