package service

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"backend-path/internal/domain"
	"backend-path/internal/interfaces"
)

type BalanceService struct {
	balanceRepo  interfaces.BalanceRepository
	auditLogRepo interfaces.AuditLogRepository
	cacheRepo    interfaces.CacheRepository
}

func NewBalanceService(
	balanceRepo interfaces.BalanceRepository,
	auditLogRepo interfaces.AuditLogRepository,
	cacheRepo interfaces.CacheRepository,
) *BalanceService {
	return &BalanceService{
		balanceRepo:  balanceRepo,
		auditLogRepo: auditLogRepo,
		cacheRepo:    cacheRepo,
	}
}

func (s *BalanceService) GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error) {
	return s.GetByUserIDAndCurrency(ctx, userID, domain.CurrencyTRY)
}

func (s *BalanceService) GetByUserIDAndCurrency(ctx context.Context, userID int64, currency domain.Currency) (*domain.Balance, error) {
	if userID <= 0 {
		return nil, domain.ErrInvalidBalanceUserID
	}

	if !currency.IsValid() {
		return nil, domain.ErrInvalidCurrency
	}

	cacheKey := balanceCacheKeyByCurrency(userID, currency)

	if s.cacheRepo != nil {
		var cachedBalance domain.Balance

		if err := s.cacheRepo.Get(ctx, cacheKey, &cachedBalance); err == nil && cachedBalance.UserID != 0 {
			if cachedBalance.Currency == "" {
				cachedBalance.Currency = currency
			}

			if err := cachedBalance.Validate(); err == nil {
				return &cachedBalance, nil
			}
		}

		if currency == domain.CurrencyTRY {
			legacyKey := balanceCacheKey(userID)

			if err := s.cacheRepo.Get(ctx, legacyKey, &cachedBalance); err == nil && cachedBalance.UserID != 0 {
				if cachedBalance.Currency == "" {
					cachedBalance.Currency = domain.CurrencyTRY
				}

				if err := cachedBalance.Validate(); err == nil {
					return &cachedBalance, nil
				}
			}
		}
	}

	balance, err := s.balanceRepo.GetByUserIDAndCurrency(ctx, userID, currency)
	if err != nil {
		return nil, err
	}

	if balance == nil {
		return nil, domain.ErrBalanceNotFound
	}

	if balance.Currency == "" {
		balance.Currency = currency
	}

	if err := balance.Validate(); err != nil {
		return nil, err
	}

	if s.cacheRepo != nil {
		_ = s.cacheRepo.Set(ctx, cacheKey, balance, balanceCacheTTL)

		if currency == domain.CurrencyTRY {
			_ = s.cacheRepo.Set(ctx, balanceCacheKey(userID), balance, balanceCacheTTL)
		}
	}

	return balance, nil
}

func (s *BalanceService) Credit(ctx context.Context, userID int64, amount float64) (*domain.Balance, error) {
	return s.CreditWithCurrency(ctx, userID, amount, domain.CurrencyTRY)
}

func (s *BalanceService) CreditWithCurrency(ctx context.Context, userID int64, amount float64, currency domain.Currency) (*domain.Balance, error) {
	if userID <= 0 {
		return nil, domain.ErrInvalidBalanceUserID
	}

	if amount <= 0 {
		return nil, domain.ErrInvalidBalanceAmount
	}

	if !currency.IsValid() {
		return nil, domain.ErrInvalidCurrency
	}

	balance, err := s.balanceRepo.GetByUserIDAndCurrency(ctx, userID, currency)
	if err != nil {
		return nil, err
	}

	if balance == nil {
		balance = &domain.Balance{
			UserID:   userID,
			Amount:   0,
			Currency: currency,
		}

		if err := balance.Validate(); err != nil {
			return nil, err
		}

		if err := s.balanceRepo.Create(ctx, balance); err != nil {
			return nil, err
		}
	}

	if balance.Currency == "" {
		balance.Currency = currency
	}

	if err := balance.Credit(amount); err != nil {
		return nil, err
	}

	if err := s.balanceRepo.Update(ctx, balance); err != nil {
		return nil, err
	}

	if s.cacheRepo != nil {
		_ = s.cacheRepo.Set(ctx, balanceCacheKeyByCurrency(userID, currency), balance, balanceCacheTTL)

		if currency == domain.CurrencyTRY {
			_ = s.cacheRepo.Set(ctx, balanceCacheKey(userID), balance, balanceCacheTTL)
		}
	}

	if s.auditLogRepo != nil {
		log := &domain.AuditLog{
			EntityType: "balance",
			EntityID:   strconv.FormatInt(userID, 10),
			Action:     "credit",
			Details:    fmt.Sprintf("credited %.2f %s, new_balance %.2f %s", amount, currency, balance.Amount, currency),
			CreatedAt:  time.Now(),
		}
		_ = s.auditLogRepo.Create(ctx, log)
	}

	return balance, nil
}

func (s *BalanceService) Debit(ctx context.Context, userID int64, amount float64) (*domain.Balance, error) {
	return s.DebitWithCurrency(ctx, userID, amount, domain.CurrencyTRY)
}

func (s *BalanceService) DebitWithCurrency(ctx context.Context, userID int64, amount float64, currency domain.Currency) (*domain.Balance, error) {
	if userID <= 0 {
		return nil, domain.ErrInvalidBalanceUserID
	}

	if amount <= 0 {
		return nil, domain.ErrInvalidBalanceAmount
	}

	if !currency.IsValid() {
		return nil, domain.ErrInvalidCurrency
	}

	balance, err := s.balanceRepo.GetByUserIDAndCurrency(ctx, userID, currency)
	if err != nil {
		return nil, err
	}

	if balance == nil {
		return nil, domain.ErrBalanceNotFound
	}

	if balance.Currency == "" {
		balance.Currency = currency
	}

	if err := balance.Debit(amount); err != nil {
		return nil, err
	}

	if err := s.balanceRepo.Update(ctx, balance); err != nil {
		return nil, err
	}

	if s.cacheRepo != nil {
		_ = s.cacheRepo.Set(ctx, balanceCacheKeyByCurrency(userID, currency), balance, balanceCacheTTL)

		if currency == domain.CurrencyTRY {
			_ = s.cacheRepo.Set(ctx, balanceCacheKey(userID), balance, balanceCacheTTL)
		}
	}

	if s.auditLogRepo != nil {
		log := &domain.AuditLog{
			EntityType: "balance",
			EntityID:   strconv.FormatInt(userID, 10),
			Action:     "debit",
			Details:    fmt.Sprintf("debited %.2f %s, new_balance %.2f %s", amount, currency, balance.Amount, currency),
			CreatedAt:  time.Now(),
		}
		_ = s.auditLogRepo.Create(ctx, log)
	}

	return balance, nil
}

func (s *BalanceService) GetHistory(ctx context.Context, userID int64) ([]*domain.AuditLog, error) {
	if userID <= 0 {
		return nil, domain.ErrInvalidBalanceUserID
	}

	if s.auditLogRepo == nil {
		return []*domain.AuditLog{}, nil
	}

	return s.auditLogRepo.ListByEntityID(ctx, "balance", strconv.FormatInt(userID, 10))
}

func (s *BalanceService) GetCurrentAmount(ctx context.Context, userID int64) (float64, error) {
	return s.GetCurrentAmountByCurrency(ctx, userID, domain.CurrencyTRY)
}

func (s *BalanceService) GetCurrentAmountByCurrency(ctx context.Context, userID int64, currency domain.Currency) (float64, error) {
	balance, err := s.GetByUserIDAndCurrency(ctx, userID, currency)
	if err != nil {
		return 0, err
	}

	return balance.GetAmount(), nil
}

const balanceCacheTTL = 5 * time.Minute

func balanceCacheKey(userID int64) string {
	return fmt.Sprintf("balance:%d", userID)
}

func balanceCacheKeyByCurrency(userID int64, currency domain.Currency) string {
	return fmt.Sprintf("balance:%d:%s", userID, currency)
}
