package domain

import "testing"

func TestCurrency_IsValid_ReturnsTrueForSupportedCurrencies(t *testing.T) {
	tests := []Currency{
		CurrencyTRY,
		CurrencyUSD,
		CurrencyEUR,
	}

	for _, currency := range tests {
		if !currency.IsValid() {
			t.Fatalf("expected currency %s to be valid", currency)
		}
	}
}

func TestCurrency_IsValid_ReturnsFalseForUnsupportedCurrency(t *testing.T) {
	currency := Currency("GBP")

	if currency.IsValid() {
		t.Fatalf("expected currency %s to be invalid", currency)
	}
}
