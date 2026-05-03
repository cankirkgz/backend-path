package service

import (
	"context"
	"strconv"
	"testing"
	"time"

	"backend-path/internal/domain"
)

type fakeBalanceRepository struct {
	data map[string]*domain.Balance
}

func newFakeBalanceRepository() *fakeBalanceRepository {
	return &fakeBalanceRepository{
		data: make(map[string]*domain.Balance),
	}
}

func (r *fakeBalanceRepository) Create(ctx context.Context, balance *domain.Balance) error {
	if balance.Currency == "" {
		balance.Currency = domain.CurrencyTRY
	}

	r.data[fakeBalanceKey(balance.UserID, balance.Currency)] = balance
	return nil
}

func (r *fakeBalanceRepository) GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error) {
	return r.GetByUserIDAndCurrency(ctx, userID, domain.CurrencyTRY)
}

func (r *fakeBalanceRepository) GetByUserIDAndCurrency(ctx context.Context, userID int64, currency domain.Currency) (*domain.Balance, error) {
	if currency == "" {
		currency = domain.CurrencyTRY
	}

	balance, ok := r.data[fakeBalanceKey(userID, currency)]
	if !ok {
		return nil, nil
	}

	return balance, nil
}

func (r *fakeBalanceRepository) Update(ctx context.Context, balance *domain.Balance) error {
	if balance.Currency == "" {
		balance.Currency = domain.CurrencyTRY
	}

	r.data[fakeBalanceKey(balance.UserID, balance.Currency)] = balance
	return nil
}

func fakeBalanceKey(userID int64, currency domain.Currency) string {
	return strconv.FormatInt(userID, 10) + ":" + string(currency)
}

func TestBalanceServiceGetByUserID_Success(t *testing.T) {
	repo := newFakeBalanceRepository()

	repo.data[fakeBalanceKey(1, domain.CurrencyTRY)] = &domain.Balance{
		UserID:        1,
		Amount:        150.50,
		LastUpdatedAt: time.Now(),
	}

	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo, nil)

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
	service := NewBalanceService(repo, auditRepo, nil)

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
	service := NewBalanceService(repo, auditRepo, nil)

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

	repo.data[fakeBalanceKey(1, domain.CurrencyTRY)] = &domain.Balance{
		UserID:        1,
		Amount:        100,
		LastUpdatedAt: time.Now(),
	}

	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo, nil)

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

	stored := repo.data[fakeBalanceKey(1, domain.CurrencyTRY)]
	if stored.Amount != 150 {
		t.Fatalf("expected stored amount=150, got %f", stored.Amount)
	}
}

func TestBalanceServiceCredit_NewBalance(t *testing.T) {
	repo := newFakeBalanceRepository()
	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo, nil)

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

	stored, ok := repo.data[fakeBalanceKey(2, domain.CurrencyTRY)]
	if !ok {
		t.Fatal("expected balance to be created in repository")
	}

	if stored.Amount != 75 {
		t.Fatalf("expected stored amount=75, got %f", stored.Amount)
	}
}

func TestBalanceServiceDebit_Success(t *testing.T) {
	repo := newFakeBalanceRepository()

	repo.data[fakeBalanceKey(1, domain.CurrencyTRY)] = &domain.Balance{
		UserID:        1,
		Amount:        200,
		LastUpdatedAt: time.Now(),
	}

	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo, nil)

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

	stored := repo.data[fakeBalanceKey(1, domain.CurrencyTRY)]
	if stored.Amount != 120 {
		t.Fatalf("expected stored amount=120, got %f", stored.Amount)
	}
}

func TestBalanceServiceDebit_InsufficientBalance(t *testing.T) {
	repo := newFakeBalanceRepository()

	repo.data[fakeBalanceKey(1, domain.CurrencyTRY)] = &domain.Balance{
		UserID:        1,
		Amount:        50,
		LastUpdatedAt: time.Now(),
	}

	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo, nil)

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

	stored := repo.data[fakeBalanceKey(1, domain.CurrencyTRY)]
	if stored.Amount != 50 {
		t.Fatalf("expected stored amount=50, got %f", stored.Amount)
	}
}

func TestBalanceServiceCredit_InvalidAmount(t *testing.T) {
	repo := newFakeBalanceRepository()
	auditRepo := newFakeAuditLogRepository()
	service := NewBalanceService(repo, auditRepo, nil)

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
	service := NewBalanceService(repo, auditRepo, nil)

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
	service := NewBalanceService(repo, auditRepo, nil)

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
	service := NewBalanceService(repo, auditRepo, nil)

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
	service := NewBalanceService(repo, auditRepo, nil)

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
	service := NewBalanceService(repo, auditRepo, nil)

	repo.data[fakeBalanceKey(1, domain.CurrencyTRY)] = &domain.Balance{
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

type fakeCacheRepository struct {
	data map[string]any
}

func newFakeCacheRepository() *fakeCacheRepository {
	return &fakeCacheRepository{
		data: make(map[string]any),
	}
}

func (r *fakeCacheRepository) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	r.data[key] = value
	return nil
}

func (r *fakeCacheRepository) Get(ctx context.Context, key string, destination any) error {
	value, ok := r.data[key]
	if !ok {
		return nil
	}

	switch dest := destination.(type) {
	case *domain.Balance:
		balanceValue, ok := value.(*domain.Balance)
		if !ok {
			return nil
		}

		*dest = *balanceValue

	case *domain.User:
		userValue, ok := value.(*domain.User)
		if !ok {
			return nil
		}

		*dest = *userValue
	}

	return nil
}

func (r *fakeCacheRepository) Delete(ctx context.Context, key string) error {
	delete(r.data, key)
	return nil
}

func TestBalanceServiceGetByUserID_UsesCacheWhenAvailable(t *testing.T) {
	repo := newFakeBalanceRepository()
	auditRepo := newFakeAuditLogRepository()
	cacheRepo := newFakeCacheRepository()

	cacheRepo.data["balance:1"] = &domain.Balance{
		UserID:        1,
		Amount:        999,
		LastUpdatedAt: time.Now(),
	}

	service := NewBalanceService(repo, auditRepo, cacheRepo)

	balance, err := service.GetByUserID(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if balance.Amount != 999 {
		t.Fatalf("expected cached amount=999, got %f", balance.Amount)
	}
}

func TestBalanceServiceGetByUserID_WritesToCacheAfterRepositoryRead(t *testing.T) {
	repo := newFakeBalanceRepository()
	auditRepo := newFakeAuditLogRepository()
	cacheRepo := newFakeCacheRepository()

	repo.data[fakeBalanceKey(1, domain.CurrencyTRY)] = &domain.Balance{
		UserID:        1,
		Amount:        150,
		LastUpdatedAt: time.Now(),
	}

	service := NewBalanceService(repo, auditRepo, cacheRepo)

	_, err := service.GetByUserID(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cachedValue, ok := cacheRepo.data["balance:1"]
	if !ok {
		t.Fatal("expected balance to be written to cache")
	}

	cachedBalance, ok := cachedValue.(*domain.Balance)
	if !ok {
		t.Fatal("expected cached value to be *domain.Balance")
	}

	if cachedBalance.Amount != 150 {
		t.Fatalf("expected cached amount=150, got %f", cachedBalance.Amount)
	}
}

func TestBalanceServiceCredit_RefreshesCache(t *testing.T) {
	repo := newFakeBalanceRepository()
	auditRepo := newFakeAuditLogRepository()
	cacheRepo := newFakeCacheRepository()

	repo.data[fakeBalanceKey(1, domain.CurrencyTRY)] = &domain.Balance{
		UserID:        1,
		Amount:        100,
		LastUpdatedAt: time.Now(),
	}

	service := NewBalanceService(repo, auditRepo, cacheRepo)

	balance, err := service.Credit(context.Background(), 1, 50)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cachedBalance := cacheRepo.data["balance:1"].(*domain.Balance)

	if cachedBalance.Amount != balance.Amount {
		t.Fatalf("expected cached amount=%f, got %f", balance.Amount, cachedBalance.Amount)
	}
}

func TestBalanceServiceDebit_RefreshesCache(t *testing.T) {
	repo := newFakeBalanceRepository()
	auditRepo := newFakeAuditLogRepository()
	cacheRepo := newFakeCacheRepository()

	repo.data[fakeBalanceKey(1, domain.CurrencyTRY)] = &domain.Balance{
		UserID:        1,
		Amount:        100,
		LastUpdatedAt: time.Now(),
	}

	service := NewBalanceService(repo, auditRepo, cacheRepo)

	balance, err := service.Debit(context.Background(), 1, 40)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	cachedBalance := cacheRepo.data["balance:1"].(*domain.Balance)

	if cachedBalance.Amount != balance.Amount {
		t.Fatalf("expected cached amount=%f, got %f", balance.Amount, cachedBalance.Amount)
	}
}

func (f *fakeBalanceService) CreditWithCurrency(ctx context.Context, userID int64, amount float64, currency domain.Currency) (*domain.Balance, error) {
	return f.Credit(ctx, userID, amount)
}

func (f *fakeBalanceService) DebitWithCurrency(ctx context.Context, userID int64, amount float64, currency domain.Currency) (*domain.Balance, error) {
	return f.Debit(ctx, userID, amount)
}

func (f *fakeBalanceService) GetByUserIDAndCurrency(ctx context.Context, userID int64, currency domain.Currency) (*domain.Balance, error) {
	return f.GetByUserID(ctx, userID)
}

func (f *fakeBalanceService) GetCurrentAmountByCurrency(ctx context.Context, userID int64, currency domain.Currency) (float64, error) {
	return f.GetCurrentAmount(ctx, userID)
}
