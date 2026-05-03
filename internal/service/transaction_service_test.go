package service

import (
	"context"
	"testing"
	"time"

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

type fakeEventStore struct {
	events []*domain.Event
}

func newFakeEventStore() *fakeEventStore {
	return &fakeEventStore{
		events: make([]*domain.Event, 0),
	}
}

func (s *fakeEventStore) Append(ctx context.Context, event *domain.Event) error {
	copied := *event
	s.events = append(s.events, &copied)
	return nil
}

func (s *fakeEventStore) ListByEntityID(ctx context.Context, entityID string) ([]*domain.Event, error) {
	result := make([]*domain.Event, 0)

	for _, event := range s.events {
		if event.EntityID == entityID {
			copied := *event
			result = append(result, &copied)
		}
	}

	return result, nil
}

func (s *fakeEventStore) ListAll(ctx context.Context) ([]*domain.Event, error) {
	result := make([]*domain.Event, 0, len(s.events))

	for _, event := range s.events {
		copied := *event
		result = append(result, &copied)
	}

	return result, nil
}

func TestTransactionServiceCredit_Success(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()

	service := NewTransactionService(txRepo, balanceService, nil)

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

	service := NewTransactionService(txRepo, balanceService, nil)

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

	service := NewTransactionService(txRepo, balanceService, nil)

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

	service := NewTransactionService(txRepo, balanceService, nil)

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

	service := NewTransactionService(txRepo, balanceService, nil)

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

	service := NewTransactionService(txRepo, balanceService, nil)

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

	service := NewTransactionService(txRepo, balanceService, nil)

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

func TestTransactionServiceCredit_RecordsEvents(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()
	eventStore := newFakeEventStore()

	service := NewTransactionService(txRepo, balanceService, eventStore)

	tx, err := service.Credit(context.Background(), 1, 100)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tx == nil {
		t.Fatal("expected transaction, got nil")
	}

	events, err := eventStore.ListAll(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if events[0].Type != domain.EventTransactionCreated {
		t.Fatalf("expected first event type %s, got %s", domain.EventTransactionCreated, events[0].Type)
	}

	if events[1].Type != domain.EventBalanceUpdated {
		t.Fatalf("expected second event type %s, got %s", domain.EventBalanceUpdated, events[1].Type)
	}

	if events[1].Data["amount"] != float64(100) {
		t.Fatalf("expected balance event amount 100, got %v", events[1].Data["amount"])
	}

	if events[1].Data["reason"] != "credit" {
		t.Fatalf("expected balance event reason credit, got %v", events[1].Data["reason"])
	}
}

func TestTransactionServiceDebit_RecordsEvents(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()
	eventStore := newFakeEventStore()

	balanceService.balances[1] = &domain.Balance{
		UserID: 1,
		Amount: 200,
	}

	service := NewTransactionService(txRepo, balanceService, eventStore)

	tx, err := service.Debit(context.Background(), 1, 80)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tx == nil {
		t.Fatal("expected transaction, got nil")
	}

	events, err := eventStore.ListAll(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(events))
	}

	if events[0].Type != domain.EventTransactionCreated {
		t.Fatalf("expected first event type %s, got %s", domain.EventTransactionCreated, events[0].Type)
	}

	if events[1].Type != domain.EventBalanceUpdated {
		t.Fatalf("expected second event type %s, got %s", domain.EventBalanceUpdated, events[1].Type)
	}

	if events[1].Data["amount"] != float64(-80) {
		t.Fatalf("expected balance event amount -80, got %v", events[1].Data["amount"])
	}

	if events[1].Data["reason"] != "debit" {
		t.Fatalf("expected balance event reason debit, got %v", events[1].Data["reason"])
	}
}

func TestTransactionServiceTransfer_RecordsEvents(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()
	eventStore := newFakeEventStore()

	balanceService.balances[1] = &domain.Balance{
		UserID: 1,
		Amount: 200,
	}

	balanceService.balances[2] = &domain.Balance{
		UserID: 2,
		Amount: 50,
	}

	service := NewTransactionService(txRepo, balanceService, eventStore)

	tx, err := service.Transfer(context.Background(), 1, 2, 80)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if tx == nil {
		t.Fatal("expected transaction, got nil")
	}

	events, err := eventStore.ListAll(context.Background())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}

	if events[0].Type != domain.EventTransactionCreated {
		t.Fatalf("expected first event type %s, got %s", domain.EventTransactionCreated, events[0].Type)
	}

	if events[1].Type != domain.EventBalanceUpdated {
		t.Fatalf("expected second event type %s, got %s", domain.EventBalanceUpdated, events[1].Type)
	}

	if events[1].Data["amount"] != float64(-80) {
		t.Fatalf("expected sender balance event amount -80, got %v", events[1].Data["amount"])
	}

	if events[1].Data["reason"] != "transfer_debit" {
		t.Fatalf("expected sender event reason transfer_debit, got %v", events[1].Data["reason"])
	}

	if events[2].Type != domain.EventBalanceUpdated {
		t.Fatalf("expected third event type %s, got %s", domain.EventBalanceUpdated, events[2].Type)
	}

	if events[2].Data["amount"] != float64(80) {
		t.Fatalf("expected receiver balance event amount 80, got %v", events[2].Data["amount"])
	}

	if events[2].Data["reason"] != "transfer_credit" {
		t.Fatalf("expected receiver event reason transfer_credit, got %v", events[2].Data["reason"])
	}
}

func TestTransactionServiceDebit_RejectsAboveMaximumAmount(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()

	balanceService.balances[1] = &domain.Balance{
		UserID: 1,
		Amount: 20000,
	}

	service := NewTransactionService(txRepo, balanceService, nil)

	tx, err := service.Debit(context.Background(), 1, 10001)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrTransactionAboveMaximumAmount {
		t.Fatalf("expected ErrTransactionAboveMaximumAmount, got %v", err)
	}

	if tx != nil {
		t.Fatalf("expected nil transaction, got %+v", tx)
	}

	if len(txRepo.data) != 0 {
		t.Fatalf("expected no transaction to be stored, got %d", len(txRepo.data))
	}

	if balanceService.balances[1].Amount != 20000 {
		t.Fatalf("expected balance amount to remain 20000, got %f", balanceService.balances[1].Amount)
	}
}

func TestTransactionServiceTransfer_RejectsDailyOutgoingLimit(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()

	balanceService.balances[1] = &domain.Balance{
		UserID: 1,
		Amount: 100000,
	}

	balanceService.balances[2] = &domain.Balance{
		UserID: 2,
		Amount: 0,
	}

	service := NewTransactionService(txRepo, balanceService, nil)

	existingTx := &domain.Transaction{
		FromUserID: 1,
		ToUserID:   2,
		Amount:     49000,
		Type:       domain.TransactionTypeTransfer,
		Status:     domain.TransactionStatusCompleted,
		CreatedAt:  time.Now(),
	}

	if err := txRepo.Create(context.Background(), existingTx); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tx, err := service.Transfer(context.Background(), 1, 2, 2000)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrDailyTransactionLimitExceeded {
		t.Fatalf("expected ErrDailyTransactionLimitExceeded, got %v", err)
	}

	if tx != nil {
		t.Fatalf("expected nil transaction, got %+v", tx)
	}

	if len(txRepo.data) != 1 {
		t.Fatalf("expected only existing transaction to remain, got %d", len(txRepo.data))
	}

	if balanceService.balances[1].Amount != 100000 {
		t.Fatalf("expected sender balance to remain 100000, got %f", balanceService.balances[1].Amount)
	}

	if balanceService.balances[2].Amount != 0 {
		t.Fatalf("expected receiver balance to remain 0, got %f", balanceService.balances[2].Amount)
	}
}

func TestTransactionServiceDebit_RejectsDailyTransactionCountLimit(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()

	balanceService.balances[1] = &domain.Balance{
		UserID: 1,
		Amount: 100000,
	}

	service := NewTransactionService(txRepo, balanceService, nil)

	for i := 0; i < 20; i++ {
		existingTx := &domain.Transaction{
			FromUserID: 1,
			Amount:     10,
			Type:       domain.TransactionTypeDebit,
			Status:     domain.TransactionStatusCompleted,
			CreatedAt:  time.Now(),
		}

		if err := txRepo.Create(context.Background(), existingTx); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
	}

	tx, err := service.Debit(context.Background(), 1, 10)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err != domain.ErrDailyTransactionCountExceeded {
		t.Fatalf("expected ErrDailyTransactionCountExceeded, got %v", err)
	}

	if tx != nil {
		t.Fatalf("expected nil transaction, got %+v", tx)
	}

	if len(txRepo.data) != 20 {
		t.Fatalf("expected existing 20 transactions only, got %d", len(txRepo.data))
	}

	if balanceService.balances[1].Amount != 100000 {
		t.Fatalf("expected balance to remain 100000, got %f", balanceService.balances[1].Amount)
	}
}

func TestTransactionServiceBatchCredit_PartialSuccess(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()

	service := NewTransactionService(txRepo, balanceService, nil)

	result, err := service.BatchCredit(context.Background(), []domain.BatchCreditItem{
		{UserID: 1, Amount: 100},
		{UserID: 2, Amount: 200},
		{UserID: 3, Amount: 0},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.TotalCount != 3 {
		t.Fatalf("expected total count 3, got %d", result.TotalCount)
	}

	if result.SuccessCount != 2 {
		t.Fatalf("expected success count 2, got %d", result.SuccessCount)
	}

	if result.FailureCount != 1 {
		t.Fatalf("expected failure count 1, got %d", result.FailureCount)
	}

	if len(result.Items) != 3 {
		t.Fatalf("expected 3 item results, got %d", len(result.Items))
	}

	if !result.Items[0].WasSuccessful {
		t.Fatal("expected first item to be successful")
	}

	if !result.Items[1].WasSuccessful {
		t.Fatal("expected second item to be successful")
	}

	if result.Items[2].WasSuccessful {
		t.Fatal("expected third item to fail")
	}

	if balanceService.balances[1].Amount != 100 {
		t.Fatalf("expected user 1 balance 100, got %f", balanceService.balances[1].Amount)
	}

	if balanceService.balances[2].Amount != 200 {
		t.Fatalf("expected user 2 balance 200, got %f", balanceService.balances[2].Amount)
	}
}

func TestTransactionServiceBatchDebit_PartialSuccess(t *testing.T) {
	txRepo := newFakeTransactionRepository()
	balanceService := newFakeTransactionBalanceService()

	balanceService.balances[1] = &domain.Balance{
		UserID: 1,
		Amount: 300,
	}

	balanceService.balances[2] = &domain.Balance{
		UserID: 2,
		Amount: 50,
	}

	service := NewTransactionService(txRepo, balanceService, nil)

	result, err := service.BatchDebit(context.Background(), []domain.BatchDebitItem{
		{UserID: 1, Amount: 100},
		{UserID: 2, Amount: 100},
		{UserID: 3, Amount: 50},
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if result.TotalCount != 3 {
		t.Fatalf("expected total count 3, got %d", result.TotalCount)
	}

	if result.SuccessCount != 1 {
		t.Fatalf("expected success count 1, got %d", result.SuccessCount)
	}

	if result.FailureCount != 2 {
		t.Fatalf("expected failure count 2, got %d", result.FailureCount)
	}

	if len(result.Items) != 3 {
		t.Fatalf("expected 3 item results, got %d", len(result.Items))
	}

	if !result.Items[0].WasSuccessful {
		t.Fatal("expected first item to be successful")
	}

	if result.Items[1].WasSuccessful {
		t.Fatal("expected second item to fail")
	}

	if result.Items[2].WasSuccessful {
		t.Fatal("expected third item to fail")
	}

	if balanceService.balances[1].Amount != 200 {
		t.Fatalf("expected user 1 balance 200, got %f", balanceService.balances[1].Amount)
	}

	if balanceService.balances[2].Amount != 50 {
		t.Fatalf("expected user 2 balance to remain 50, got %f", balanceService.balances[2].Amount)
	}
}
