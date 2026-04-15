package service

import (
	"context"
	"testing"

	"backend-path/internal/domain"
)

type fakeTransactionRepository struct {
	data   map[int64]*domain.Transaction
	nextID int64
}

func newFakeTransactionRepository() *fakeTransactionRepository {
	return &fakeTransactionRepository{
		data:   make(map[int64]*domain.Transaction),
		nextID: 1,
	}
}

func (r *fakeTransactionRepository) Create(ctx context.Context, tx *domain.Transaction) error {
	tx.ID = r.nextID
	r.nextID++
	r.data[tx.ID] = tx
	return nil
}

func (r *fakeTransactionRepository) GetByID(ctx context.Context, id int64) (*domain.Transaction, error) {
	tx, ok := r.data[id]
	if !ok {
		return nil, nil
	}
	return tx, nil
}

func (r *fakeTransactionRepository) ListByUserID(ctx context.Context, userID int64) ([]*domain.Transaction, error) {
	var result []*domain.Transaction

	for _, tx := range r.data {
		if tx.FromUserID == userID || tx.ToUserID == userID {
			result = append(result, tx)
		}
	}

	return result, nil
}

func (r *fakeTransactionRepository) UpdateStatus(ctx context.Context, id int64, status domain.TransactionStatus) error {
	tx, ok := r.data[id]
	if !ok {
		return nil
	}

	tx.Status = status
	return nil
}

type fakeBalanceService struct {
	balances map[int64]*domain.Balance
}

func newFakeTransactionBalanceService() *fakeBalanceService {
	return &fakeBalanceService{
		balances: make(map[int64]*domain.Balance),
	}
}

func (s *fakeBalanceService) GetByUserID(ctx context.Context, userID int64) (*domain.Balance, error) {
	balance, ok := s.balances[userID]
	if !ok {
		return nil, domain.ErrBalanceNotFound
	}
	return balance, nil
}

func (s *fakeBalanceService) Credit(ctx context.Context, userID int64, amount float64) (*domain.Balance, error) {
	if userID <= 0 {
		return nil, domain.ErrInvalidBalanceUserID
	}

	if amount <= 0 {
		return nil, domain.ErrInvalidBalanceAmount
	}

	balance, ok := s.balances[userID]
	if !ok {
		balance = &domain.Balance{
			UserID: userID,
			Amount: 0,
		}
		s.balances[userID] = balance
	}

	if err := balance.Credit(amount); err != nil {
		return nil, err
	}

	return balance, nil
}

func (s *fakeBalanceService) Debit(ctx context.Context, userID int64, amount float64) (*domain.Balance, error) {
	if userID <= 0 {
		return nil, domain.ErrInvalidBalanceUserID
	}

	if amount <= 0 {
		return nil, domain.ErrInvalidBalanceAmount
	}

	balance, ok := s.balances[userID]
	if !ok {
		return nil, domain.ErrBalanceNotFound
	}

	if err := balance.Debit(amount); err != nil {
		return nil, err
	}

	return balance, nil
}

func (s *fakeBalanceService) GetHistory(ctx context.Context, userID int64) ([]*domain.AuditLog, error) {
	return []*domain.AuditLog{}, nil
}

func (s *fakeBalanceService) GetCurrentAmount(ctx context.Context, userID int64) (float64, error) {
	balance, ok := s.balances[userID]
	if !ok {
		return 0, domain.ErrBalanceNotFound
	}

	return balance.GetAmount(), nil
}

func TestTransactionServiceCredit_Success(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()

	service := NewTransactionService(txRepo, balanceService)

	tx, err := service.Credit(context.Background(), 1, 100)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tx == nil {
		t.Fatal("expected transaction, got nil")
	}

	if tx.ID == 0 {
		t.Fatalf("expected transaction ID to be set, got %d", tx.ID)
	}

	if tx.Type != domain.TransactionTypeCredit {
		t.Fatalf("expected type=%s, got %s", domain.TransactionTypeCredit, tx.Type)
	}

	if tx.Status != domain.TransactionStatusCompleted {
		t.Fatalf("expected status=%s, got %s", domain.TransactionStatusCompleted, tx.Status)
	}

	if tx.ToUserID != 1 {
		t.Fatalf("expected ToUserID=1, got %d", tx.ToUserID)
	}

	if tx.Amount != 100 {
		t.Fatalf("expected amount=100, got %f", tx.Amount)
	}

	storedTx, ok := txRepo.data[tx.ID]
	if !ok {
		t.Fatal("expected transaction to be stored in repository")
	}

	if storedTx.Status != domain.TransactionStatusCompleted {
		t.Fatalf("expected stored status=%s, got %s", domain.TransactionStatusCompleted, storedTx.Status)
	}

	balance, ok := balanceService.balances[1]
	if !ok {
		t.Fatal("expected balance to be updated")
	}

	if balance.Amount != 100 {
		t.Fatalf("expected balance amount=100, got %f", balance.Amount)
	}
}

func TestTransactionServiceCredit_InvalidAmount(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()

	service := NewTransactionService(txRepo, balanceService)

	tx, err := service.Credit(context.Background(), 1, 0)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrInvalidTransactionAmount {
		t.Fatalf("expected ErrInvalidTransactionAmount, got %v", err)
	}

	if tx != nil {
		t.Fatalf("expected nil transaction, got %+v", tx)
	}

	if len(txRepo.data) != 0 {
		t.Fatalf("expected no transaction to be stored, got %d", len(txRepo.data))
	}

	if len(balanceService.balances) != 0 {
		t.Fatalf("expected no balance update, got %d", len(balanceService.balances))
	}
}

func TestTransactionServiceDebit_Success(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()

	balanceService.balances[1] = &domain.Balance{
		UserID: 1,
		Amount: 200,
	}

	service := NewTransactionService(txRepo, balanceService)

	tx, err := service.Debit(context.Background(), 1, 80)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tx == nil {
		t.Fatal("expected transaction, got nil")
	}

	if tx.ID == 0 {
		t.Fatalf("expected transaction ID to be set, got %d", tx.ID)
	}

	if tx.Type != domain.TransactionTypeDebit {
		t.Fatalf("expected type=%s, got %s", domain.TransactionTypeDebit, tx.Type)
	}

	if tx.Status != domain.TransactionStatusCompleted {
		t.Fatalf("expected status=%s, got %s", domain.TransactionStatusCompleted, tx.Status)
	}

	if tx.FromUserID != 1 {
		t.Fatalf("expected FromUserID=1, got %d", tx.FromUserID)
	}

	if tx.Amount != 80 {
		t.Fatalf("expected amount=80, got %f", tx.Amount)
	}

	storedTx, ok := txRepo.data[tx.ID]
	if !ok {
		t.Fatal("expected transaction to be stored in repository")
	}

	if storedTx.Status != domain.TransactionStatusCompleted {
		t.Fatalf("expected stored status=%s, got %s", domain.TransactionStatusCompleted, storedTx.Status)
	}

	balance, ok := balanceService.balances[1]
	if !ok {
		t.Fatal("expected balance to exist")
	}

	if balance.Amount != 120 {
		t.Fatalf("expected balance amount=120, got %f", balance.Amount)
	}
}

func TestTransactionServiceDebit_InsufficientBalance(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()

	balanceService.balances[1] = &domain.Balance{
		UserID: 1,
		Amount: 50,
	}

	service := NewTransactionService(txRepo, balanceService)

	tx, err := service.Debit(context.Background(), 1, 80)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrInsufficientBalance {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}

	if tx != nil {
		t.Fatalf("expected nil transaction, got %+v", tx)
	}

	if len(txRepo.data) != 1 {
		t.Fatalf("expected 1 transaction in repository, got %d", len(txRepo.data))
	}

	var storedTx *domain.Transaction
	for _, item := range txRepo.data {
		storedTx = item
		break
	}

	if storedTx == nil {
		t.Fatal("expected stored transaction, got nil")
	}

	if storedTx.Status != domain.TransactionStatusFailed {
		t.Fatalf("expected stored status=%s, got %s", domain.TransactionStatusFailed, storedTx.Status)
	}

	balance, ok := balanceService.balances[1]
	if !ok {
		t.Fatal("expected balance to exist")
	}

	if balance.Amount != 50 {
		t.Fatalf("expected balance amount=50, got %f", balance.Amount)
	}
}

func TestTransactionServiceTransfer_Success(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()

	balanceService.balances[1] = &domain.Balance{
		UserID: 1,
		Amount: 200,
	}

	balanceService.balances[2] = &domain.Balance{
		UserID: 2,
		Amount: 50,
	}

	service := NewTransactionService(txRepo, balanceService)

	tx, err := service.Transfer(context.Background(), 1, 2, 80)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tx == nil {
		t.Fatal("expected transaction, got nil")
	}

	if tx.ID == 0 {
		t.Fatalf("expected transaction ID to be set, got %d", tx.ID)
	}

	if tx.Type != domain.TransactionTypeTransfer {
		t.Fatalf("expected type=%s, got %s", domain.TransactionTypeTransfer, tx.Type)
	}

	if tx.Status != domain.TransactionStatusCompleted {
		t.Fatalf("expected status=%s, got %s", domain.TransactionStatusCompleted, tx.Status)
	}

	if tx.FromUserID != 1 {
		t.Fatalf("expected FromUserID=1, got %d", tx.FromUserID)
	}

	if tx.ToUserID != 2 {
		t.Fatalf("expected ToUserID=2, got %d", tx.ToUserID)
	}

	if tx.Amount != 80 {
		t.Fatalf("expected amount=80, got %f", tx.Amount)
	}

	storedTx, ok := txRepo.data[tx.ID]
	if !ok {
		t.Fatal("expected transaction to be stored in repository")
	}

	if storedTx.Status != domain.TransactionStatusCompleted {
		t.Fatalf("expected stored status=%s, got %s", domain.TransactionStatusCompleted, storedTx.Status)
	}

	fromBalance, ok := balanceService.balances[1]
	if !ok {
		t.Fatal("expected sender balance to exist")
	}

	if fromBalance.Amount != 120 {
		t.Fatalf("expected sender balance amount=120, got %f", fromBalance.Amount)
	}

	toBalance, ok := balanceService.balances[2]
	if !ok {
		t.Fatal("expected receiver balance to exist")
	}

	if toBalance.Amount != 130 {
		t.Fatalf("expected receiver balance amount=130, got %f", toBalance.Amount)
	}
}

func TestTransactionServiceTransfer_InsufficientBalance(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()

	balanceService.balances[1] = &domain.Balance{
		UserID: 1,
		Amount: 50,
	}

	balanceService.balances[2] = &domain.Balance{
		UserID: 2,
		Amount: 20,
	}

	service := NewTransactionService(txRepo, balanceService)

	tx, err := service.Transfer(context.Background(), 1, 2, 80)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrInsufficientBalance {
		t.Fatalf("expected ErrInsufficientBalance, got %v", err)
	}

	if tx != nil {
		t.Fatalf("expected nil transaction, got %+v", tx)
	}

	if len(txRepo.data) != 1 {
		t.Fatalf("expected 1 transaction in repository, got %d", len(txRepo.data))
	}

	var storedTx *domain.Transaction
	for _, item := range txRepo.data {
		storedTx = item
		break
	}

	if storedTx == nil {
		t.Fatal("expected stored transaction, got nil")
	}

	if storedTx.Status != domain.TransactionStatusFailed {
		t.Fatalf("expected stored status=%s, got %s", domain.TransactionStatusFailed, storedTx.Status)
	}

	fromBalance, ok := balanceService.balances[1]
	if !ok {
		t.Fatal("expected sender balance to exist")
	}

	if fromBalance.Amount != 50 {
		t.Fatalf("expected sender balance amount=50, got %f", fromBalance.Amount)
	}

	toBalance, ok := balanceService.balances[2]
	if !ok {
		t.Fatal("expected receiver balance to exist")
	}

	if toBalance.Amount != 20 {
		t.Fatalf("expected receiver balance amount=20, got %f", toBalance.Amount)
	}
}

func TestTransactionServiceTransfer_SameSenderReceiver(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()

	balanceService.balances[1] = &domain.Balance{
		UserID: 1,
		Amount: 100,
	}

	service := NewTransactionService(txRepo, balanceService)

	tx, err := service.Transfer(context.Background(), 1, 1, 50)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrSameSenderReceiver {
		t.Fatalf("expected ErrSameSenderReceiver, got %v", err)
	}

	if tx != nil {
		t.Fatalf("expected nil transaction, got %+v", tx)
	}

	if len(txRepo.data) != 0 {
		t.Fatalf("expected no transaction to be stored, got %d", len(txRepo.data))
	}

	balance, ok := balanceService.balances[1]
	if !ok {
		t.Fatal("expected balance to exist")
	}

	if balance.Amount != 100 {
		t.Fatalf("expected balance amount=100, got %f", balance.Amount)
	}
}
