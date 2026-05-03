package domain

import "time"

type EventType string

const (
	EventTransactionCreated EventType = "transaction_created"
	EventBalanceUpdated     EventType = "balance_updated"
)

type Event struct {
	ID        int64
	Type      EventType
	EntityID  string // user_id ya da transaction_id
	Data      map[string]interface{}
	CreatedAt time.Time
}
