package domain

import "testing"

func TestTransaction_Validate_ReturnsErrorForInvalidCurrency(t *testing.T) {
	tx := &Transaction{
		ToUserID: 1,
		Amount:   100,
		Currency: Currency("GBP"),
		Type:     TransactionTypeCredit,
		Status:   TransactionStatusPending,
	}

	err := tx.Validate()
	if err != ErrInvalidCurrency {
		t.Fatalf("expected %v, got %v", ErrInvalidCurrency, err)
	}
}

func TestTransaction_Validate_AllowsValidCurrency(t *testing.T) {
	tx := &Transaction{
		ToUserID: 1,
		Amount:   100,
		Currency: CurrencyTRY,
		Type:     TransactionTypeCredit,
		Status:   TransactionStatusPending,
	}

	err := tx.Validate()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
