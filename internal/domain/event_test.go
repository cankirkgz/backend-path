package domain

import "testing"

func TestEventCreation(t *testing.T) {
	event := Event{
		Type:     EventTransactionCreated,
		EntityID: "1",
		Data: map[string]interface{}{
			"amount": 100,
		},
	}

	if event.Type != EventTransactionCreated {
		t.Fatalf("expected event type %s, got %s", EventTransactionCreated, event.Type)
	}

	if event.EntityID != "1" {
		t.Fatalf("expected entity id 1, got %s", event.EntityID)
	}
}
