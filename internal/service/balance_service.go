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
}

func NewBalanceService(
	balanceRepo interfaces.BalanceRepository,
	auditLogRepo interfaces.AuditLogRepository,
) *BalanceService {
	return &BalanceService{
		balanceRepo:  balanceRepo,
		auditLogRepo: auditLogRepo,
	}
}
func (s *BalanceService) GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error) {
	if userID <= 0 {
		return nil, domain.ErrInvalidBalanceUserID
	}

	balance, err := s.balanceRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if balance == nil {
		return nil, domain.ErrBalanceNotFound
	}

	if err := balance.Validate(); err != nil {
		return nil, err
	}

	return balance, nil
}

func (s *BalanceService) Credit(ctx context.Context, userID int64, amount float64) (*domain.Balance, error) {
	if userID <= 0 {
		return nil, domain.ErrInvalidBalanceUserID
	}

	if amount <= 0 {
		return nil, domain.ErrInvalidBalanceAmount
	}

	balance, err := s.balanceRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if balance == nil {
		balance = &domain.Balance{
			UserID: userID,
			Amount: 0,
		}

		if err := balance.Validate(); err != nil {
			return nil, err
		}

		if err := s.balanceRepo.Create(ctx, balance); err != nil {
			return nil, err
		}
	}

	if err := balance.Credit(amount); err != nil {
		return nil, err
	}

	if err := s.balanceRepo.Update(ctx, balance); err != nil {
		return nil, err
	}

	if s.auditLogRepo != nil {
		log := &domain.AuditLog{
			EntityType: "balance",
			EntityID:   strconv.FormatInt(userID, 10),
			Action:     "credit",
			Details:    fmt.Sprintf("credited %.2f, new_balance %.2f", amount, balance.Amount),
			CreatedAt:  time.Now(),
		}
		_ = s.auditLogRepo.Create(ctx, log)
	}

	return balance, nil
}

func (s *BalanceService) Debit(ctx context.Context, userID int64, amount float64) (*domain.Balance, error) {
	if userID <= 0 {
		return nil, domain.ErrInvalidBalanceUserID
	}

	if amount <= 0 {
		return nil, domain.ErrInvalidBalanceAmount
	}

	balance, err := s.balanceRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	if balance == nil {
		return nil, domain.ErrBalanceNotFound
	}

	if err := balance.Debit(amount); err != nil {
		return nil, err
	}

	if err := s.balanceRepo.Update(ctx, balance); err != nil {
		return nil, err
	}

	if s.auditLogRepo != nil {
		log := &domain.AuditLog{
			EntityType: "balance",
			EntityID:   strconv.FormatInt(userID, 10),
			Action:     "debit",
			Details:    fmt.Sprintf("debited %.2f, new_balance %.2f", amount, balance.Amount),
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
	balance, err := s.GetByUserID(ctx, userID)
	if err != nil {
		return 0, err
	}

	return balance.GetAmount(), nil
}
