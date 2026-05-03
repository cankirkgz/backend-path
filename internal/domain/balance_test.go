package domain

import "testing"

func TestBalance_Validate_ReturnsErrorForInvalidCurrency(t *testing.T) {
	balance := &Balance{
		UserID:   1,
		Amount:   100,
		Currency: Currency("GBP"),
	}

	err := balance.Validate()
	if err != ErrInvalidCurrency {
		t.Fatalf("expected %v, got %v", ErrInvalidCurrency, err)
	}
}

func TestBalance_Validate_AllowsValidCurrency(t *testing.T) {
	balance := &Balance{
		UserID:   1,
		Amount:   100,
		Currency: CurrencyTRY,
	}

	err := balance.Validate()
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
}
