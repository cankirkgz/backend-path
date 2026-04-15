package service

import (
	"context"
	"testing"
	"time"

	"backend-path/internal/domain"
)

type fakeBalanceRepository struct {
	data map[int64]*domain.Balance
}

func newFakeBalanceRepository() *fakeBalanceRepository {
	return &fakeBalanceRepository{
		data: make(map[int64]*domain.Balance),
	}
}

func (r *fakeBalanceRepository) Create(ctx context.Context, balance *domain.Balance) error {
	r.data[balance.UserID] = balance
	return nil
}

func (r *fakeBalanceRepository) GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error) {
	balance, ok := r.data[userID]
	if !ok {
		return nil, nil
	}
	return balance, nil
}

func (r *fakeBalanceRepository) Update(ctx context.Context, balance *domain.Balance) error {
	r.data[balance.UserID] = balance
	return nil
}

func TestBalanceServiceGetByUserID_Success(t *testing.T) {
	repo := newFakeBalanceRepository()

	repo.data[1] = &domain.Balance{
		UserID:        1,
		Amount:        150.50,
		LastUpdatedAt: time.Now(),
	}

	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo)

	balance, err := service.GetByUserID(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if balance == nil {
		t.Fatal("expected balance, got nil")
	}

	if balance.UserID != 1 {
		t.Fatalf("expected userID=1, got %d", balance.UserID)
	}

	if balance.Amount != 150.50 {
		t.Fatalf("expected amount=150.50, got %f", balance.Amount)
	}
}

func TestBalanceServiceGetByUserID_InvalidUserID(t *testing.T) {
	repo := newFakeBalanceRepository()
	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo)

	balance, err := service.GetByUserID(context.Background(), 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrInvalidBalanceUserID {
		t.Fatalf("expected ErrInvalidBalanceUserID, got %v", err)
	}

	if balance != nil {
		t.Fatalf("expected nil balance, got %+v", balance)
	}
}

func TestBalanceServiceGetByUserID_BalanceNotFound(t *testing.T) {
	repo := newFakeBalanceRepository()
	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo)

	balance, err := service.GetByUserID(context.Background(), 99)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrBalanceNotFound {
		t.Fatalf("expected ErrBalanceNotFound, got %v", err)
	}

	if balance != nil {
		t.Fatalf("expected nil balance, got %+v", balance)
	}
}

func TestBalanceServiceCredit_ExistingBalance(t *testing.T) {
	repo := newFakeBalanceRepository()

	repo.data[1] = &domain.Balance{
		UserID:        1,
		Amount:        100,
		LastUpdatedAt: time.Now(),
	}

	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo)

	balance, err := service.Credit(context.Background(), 1, 50)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if balance == nil {
		t.Fatal("expected balance, got nil")
	}

	if balance.Amount != 150 {
		t.Fatalf("expected amount=150, got %f", balance.Amount)
	}

	stored := repo.data[1]
	if stored.Amount != 150 {
		t.Fatalf("expected stored amount=150, got %f", stored.Amount)
	}
}

func TestBalanceServiceCredit_NewBalance(t *testing.T) {
	repo := newFakeBalanceRepository()
	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo)

	balance, err := service.Credit(context.Background(), 2, 75)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if balance == nil {
		t.Fatal("expected balance, got nil")
	}

	if balance.UserID != 2 {
		t.Fatalf("expected userID=2, got %d", balance.UserID)
	}

	if balance.Amount != 75 {
		t.Fatalf("expected amount=75, got %f", balance.Amount)
	}

	stored, ok := repo.data[2]
	if !ok {
		t.Fatal("expected balance to be created in repository")
	}

	if stored.Amount != 75 {
		t.Fatalf("expected stored amount=75, got %f", stored.Amount)
	}
}

func TestBalanceServiceDebit_Success(t *testing.T) {
	repo := newFakeBalanceRepository()

	repo.data[1] = &domain.Balance{
		UserID:        1,
		Amount:        200,
		LastUpdatedAt: time.Now(),
	}

	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo)

	balance, err := service.Debit(context.Background(), 1, 80)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if balance == nil {
		t.Fatal("expected balance, got nil")
	}

	if balance.Amount != 120 {
		t.Fatalf("expected amount=120, got %f", balance.Amount)
	}

	stored := repo.data[1]
	if stored.Amount != 120 {
		t.Fatalf("expected stored amount=120, got %f", stored.Amount)
	}
}

func TestBalanceServiceDebit_InsufficientBalance(t *testing.T) {
	repo := newFakeBalanceRepository()

	repo.data[1] = &domain.Balance{
		UserID:        1,
		Amount:        50,
		LastUpdatedAt: time.Now(),
	}

	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo)

	balance, err := service.Debit(context.Background(), 1, 80)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrInsufficientBalance {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}

	if balance != nil {
		t.Fatalf("expected nil balance, got %+v", balance)
	}

	stored := repo.data[1]
	if stored.Amount != 50 {
		t.Fatalf("expected stored amount=50, got %f", stored.Amount)
	}
}

func TestBalanceServiceCredit_InvalidAmount(t *testing.T) {
	repo := newFakeBalanceRepository()
	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo)

	balance, err := service.Credit(context.Background(), 1, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrInvalidBalanceAmount {
		t.Fatalf("expected ErrInvalidBalanceAmount, got %v", err)
	}

	if balance != nil {
		t.Fatalf("expected nil balance, got %+v", balance)
	}
}

func TestBalanceServiceDebit_InvalidAmount(t *testing.T) {
	repo := newFakeBalanceRepository()
	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo)

	balance, err := service.Debit(context.Background(), 1, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrInvalidBalanceAmount {
		t.Fatalf("expected ErrInvalidBalanceAmount, got %v", err)
	}

	if balance != nil {
		t.Fatalf("expected nil balance, got %+v", balance)
	}
}

func TestBalanceServiceDebit_BalanceNotFound(t *testing.T) {
	repo := newFakeBalanceRepository()
	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo)

	balance, err := service.Debit(context.Background(), 42, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrBalanceNotFound {
		t.Fatalf("expected ErrBalanceNotFound, got %v", err)
	}

	if balance != nil {
		t.Fatalf("expected nil balance, got %+v", balance)
	}
}

type fakeAuditLogRepository struct {
	logs []*domain.AuditLog
}

func newFakeAuditLogRepository() *fakeAuditLogRepository {
	return &fakeAuditLogRepository{
		logs: []*domain.AuditLog{},
	}
}

func (r *fakeAuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	r.logs = append(r.logs, log)
	return nil
}

func (r *fakeAuditLogRepository) ListByEntityID(ctx context.Context, entityType, entityID string) ([]*domain.AuditLog, error) {
	var result []*domain.AuditLog
	for _, log := range r.logs {
		if log.EntityType == entityType && log.EntityID == entityID {
			result = append(result, log)
		}
	}
	return result, nil
}

func TestBalanceServiceCredit_CreatesAuditLog(t *testing.T) {
	repo := newFakeBalanceRepository()
	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo)

	_, err := service.Credit(context.Background(), 1, 50)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(auditRepo.logs) != 1 {
		t.Fatalf("expected 1 audit log, got %d", len(auditRepo.logs))
	}

	if auditRepo.logs[0].Action != "credit" {
		t.Fatalf("expected action=credit, got %s", auditRepo.logs[0].Action)
	}
}

func TestBalanceServiceGetHistory_Success(t *testing.T) {
	repo := newFakeBalanceRepository()
	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo)

	auditRepo.logs = append(auditRepo.logs,
		&domain.AuditLog{EntityType: "balance", EntityID: "1", Action: "credit"},
		&domain.AuditLog{EntityType: "balance", EntityID: "1", Action: "debit"},
	)

	history, err := service.GetHistory(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(history) != 2 {
		t.Fatalf("expected history length=2, got %d", len(history))
	}
}

func TestBalanceServiceGetCurrentAmount_Success(t *testing.T) {
	repo := newFakeBalanceRepository()
	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo)

	repo.data[1] = &domain.Balance{
		UserID:        1,
		Amount:        245.75,
		LastUpdatedAt: time.Now(),
	}

	amount, err := service.GetCurrentAmount(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if amount != 245.75 {
		t.Fatalf("expected amount=245.75, got %f", amount)
	}
}
