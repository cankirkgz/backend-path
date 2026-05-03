package domain

import (
	"sync"
	"time"
)

type Balance struct {
	UserID        int64     `json:"user_id"`
	Amount        float64   `json:"amount"`
	Currency      Currency  `json:"currency"`
	LastUpdatedAt time.Time `json:"last_updated_at"`

	mu sync.RWMutex
}

func (b *Balance) Validate() error {
	if b.UserID <= 0 {
		return ErrInvalidBalanceUserID
	}

	if b.Amount < 0 {
		return ErrInvalidBalanceAmount
	}

	if !b.Currency.IsValid() {
		return ErrInvalidCurrency
	}

	return nil
}

func (b *Balance) GetAmount() float64 {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.Amount
}

func (b *Balance) Credit(amount float64) error {
	if amount <= 0 {
		return ErrInvalidBalanceAmount
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	b.Amount += amount
	b.LastUpdatedAt = time.Now()

	return nil
}

func (b *Balance) Debit(amount float64) error {
	if amount <= 0 {
		return ErrInvalidBalanceAmount
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	if b.Amount < amount {
		return ErrInsufficientBalance
	}

	b.Amount -= amount
	b.LastUpdatedAt = time.Now()

	return nil
}
