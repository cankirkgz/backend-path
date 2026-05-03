package memory

import (
	"context"
	"testing"

	"backend-path/internal/domain"
)

func TestEventStoreRepository_AppendAndListByEntityID(t *testing.T) {
	repo := NewEventStoreRepository()
	ctx := context.Background()

	err := repo.Append(ctx, &domain.Event{
		Type:     domain.EventTransactionCreated,
		EntityID: "transaction-1",
		Data: map[string]interface{}{
			"amount": 100,
		},
	})
	if err != nil {
		t.Fatalf("unexpected append error: %v", err)
	}

	events, err := repo.ListByEntityID(ctx, "transaction-1")
	if err != nil {
		t.Fatalf("unexpected list error: %v", err)
	}

	if len(events) != 1 {
		t.Fatalf("expected 1 event, got %d", len(events))
	}

	if events[0].ID != 1 {
		t.Fatalf("expected event id 1, got %d", events[0].ID)
	}

	if events[0].Type != domain.EventTransactionCreated {
		t.Fatalf("expected event type %s, got %s", domain.EventTransactionCreated, events[0].Type)
	}

	if events[0].CreatedAt.IsZero() {
		t.Fatal("expected created at to be set")
	}
}

func TestEventStoreRepository_ListAll(t *testing.T) {
	repo := NewEventStoreRepository()
	ctx := context.Background()

	_ = repo.Append(ctx, &domain.Event{
		Type:     domain.EventTransactionCreated,
		EntityID: "transaction-1",
	})

	_ = repo.Append(ctx, &domain.Event{
		Type:     domain.EventBalanceUpdated,
		EntityID: "user-1",
	})

	events, err := repo.ListAll(ctx)
	if err != nil {
		t.Fatalf("unexpected list all error: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}
}
