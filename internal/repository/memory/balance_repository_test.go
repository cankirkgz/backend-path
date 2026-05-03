package memory

import (
	"context"
	"testing"

	"backend-path/internal/domain"
)

func TestBalanceRepository_StoresBalancesByUserIDAndCurrency(t *testing.T) {
	ctx := context.Background()
	repo := NewBalanceRepository()

	tryBalance := &domain.Balance{
		UserID:   1,
		Amount:   100,
		Currency: domain.CurrencyTRY,
	}

	usdBalance := &domain.Balance{
		UserID:   1,
		Amount:   50,
		Currency: domain.CurrencyUSD,
	}

	if err := repo.Create(ctx, tryBalance); err != nil {
		t.Fatalf("unexpected error creating TRY balance: %v", err)
	}

	if err := repo.Create(ctx, usdBalance); err != nil {
		t.Fatalf("unexpected error creating USD balance: %v", err)
	}

	gotTRY, err := repo.GetByUserIDAndCurrency(ctx, 1, domain.CurrencyTRY)
	if err != nil {
		t.Fatalf("unexpected error getting TRY balance: %v", err)
	}

	gotUSD, err := repo.GetByUserIDAndCurrency(ctx, 1, domain.CurrencyUSD)
	if err != nil {
		t.Fatalf("unexpected error getting USD balance: %v", err)
	}

	if gotTRY.Amount != 100 {
		t.Fatalf("expected TRY amount 100, got %.2f", gotTRY.Amount)
	}

	if gotUSD.Amount != 50 {
		t.Fatalf("expected USD amount 50, got %.2f", gotUSD.Amount)
	}
}

func TestBalanceRepository_GetByUserID_ReturnsTRYBalanceForBackwardCompatibility(t *testing.T) {
	ctx := context.Background()
	repo := NewBalanceRepository()

	balance := &domain.Balance{
		UserID:   1,
		Amount:   100,
		Currency: domain.CurrencyTRY,
	}

	if err := repo.Create(ctx, balance); err != nil {
		t.Fatalf("unexpected error creating balance: %v", err)
	}

	got, err := repo.GetByUserID(ctx, 1)
	if err != nil {
		t.Fatalf("unexpected error getting balance: %v", err)
	}

	if got == nil {
		t.Fatalf("expected balance, got nil")
	}

	if got.Currency != domain.CurrencyTRY {
		t.Fatalf("expected currency %s, got %s", domain.CurrencyTRY, got.Currency)
	}

	if got.Amount != 100 {
		t.Fatalf("expected amount 100, got %.2f", got.Amount)
	}
}
