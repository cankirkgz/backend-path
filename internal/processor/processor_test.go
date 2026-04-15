package processor

import (
	"context"
	"testing"
	"time"

	"backend-path/internal/domain"
)

func TestProcessorProcessesJobs(t *testing.T) {
	p := NewProcessor(2, 10)
	if err := p.Start(); err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}

	validTx := &domain.Transaction{
		ID:        1,
		ToUserID:  1,
		Amount:    100,
		Type:      domain.TransactionTypeCredit,
		Status:    domain.TransactionStatusPending,
		CreatedAt: time.Now(),
	}

	invalidTx := &domain.Transaction{
		ID:        2,
		ToUserID:  1,
		Amount:    -50,
		Type:      domain.TransactionTypeCredit,
		Status:    domain.TransactionStatusPending,
		CreatedAt: time.Now(),
	}

	if err := p.Submit(Job{
		ID:          "job-1",
		Transaction: validTx,
		Ctx:         context.Background(),
	}); err != nil {
		t.Fatalf("unexpected submit error for valid job: %v", err)
	}

	if err := p.Submit(Job{
		ID:          "job-2",
		Transaction: invalidTx,
		Ctx:         context.Background(),
	}); err != nil {
		t.Fatalf("unexpected submit error for invalid job: %v", err)
	}

	p.Stop()

	if got := p.Stats().Processed(); got != 1 {
		t.Fatalf("expected processed=1, got %d", got)
	}

	if got := p.Stats().Failed(); got != 1 {
		t.Fatalf("expected failed=1, got %d", got)
	}
}

func TestProcessorSkipsCancelledJob(t *testing.T) {
	p := NewProcessor(1, 10)

	if err := p.Start(); err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	tx := &domain.Transaction{
		ID:        3,
		ToUserID:  1,
		Amount:    100,
		Type:      domain.TransactionTypeCredit,
		Status:    domain.TransactionStatusPending,
		CreatedAt: time.Now(),
	}

	if err := p.Submit(Job{
		ID:          "job-cancelled",
		Transaction: tx,
		Ctx:         ctx,
	}); err != nil {
		t.Fatalf("unexpected submit error: %v", err)
	}

	p.Stop()

	if got := p.Stats().Cancelled(); got != 1 {
		t.Fatalf("expected cancelled=1, got %d", got)
	}

	if got := p.Stats().Processed(); got != 0 {
		t.Fatalf("expected processed=0, got %d", got)
	}

	if got := p.Stats().Failed(); got != 0 {
		t.Fatalf("expected failed=0, got %d", got)
	}
}

func TestProcessorSubmitBatch(t *testing.T) {
	p := NewProcessor(2, 10)

	if err := p.Start(); err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}

	jobs := []Job{
		{
			ID: "job-1",
			Transaction: &domain.Transaction{
				ID:        10,
				ToUserID:  1,
				Amount:    100,
				Type:      domain.TransactionTypeCredit,
				Status:    domain.TransactionStatusPending,
				CreatedAt: time.Now(),
			},
			Ctx: context.Background(),
		},
		{
			ID: "job-2",
			Transaction: &domain.Transaction{
				ID:        11,
				ToUserID:  2,
				Amount:    200,
				Type:      domain.TransactionTypeCredit,
				Status:    domain.TransactionStatusPending,
				CreatedAt: time.Now(),
			},
			Ctx: context.Background(),
		},
		{
			ID: "job-3",
			Transaction: &domain.Transaction{
				ID:        12,
				ToUserID:  3,
				Amount:    -25, // ❌ invalid
				Type:      domain.TransactionTypeCredit,
				Status:    domain.TransactionStatusPending,
				CreatedAt: time.Now(),
			},
			Ctx: context.Background(),
		},
	}

	if err := p.SubmitBatch(jobs); err != nil {
		t.Fatalf("unexpected batch submit error: %v", err)
	}

	p.Stop()

	if got := p.Stats().Processed(); got != 2 {
		t.Fatalf("expected processed=2, got %d", got)
	}

	if got := p.Stats().Failed(); got != 1 {
		t.Fatalf("expected failed=1, got %d", got)
	}

	if got := p.Stats().Cancelled(); got != 0 {
		t.Fatalf("expected cancelled=0, got %d", got)
	}
}

func TestProcessorSnapshot(t *testing.T) {
	p := NewProcessor(1, 10)

	if err := p.Start(); err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}

	tx := &domain.Transaction{
		ID:        100,
		ToUserID:  1,
		Amount:    100,
		Type:      domain.TransactionTypeCredit,
		Status:    domain.TransactionStatusPending,
		CreatedAt: time.Now(),
	}

	if err := p.Submit(Job{
		ID:          "job-snapshot",
		Transaction: tx,
		Ctx:         context.Background(),
	}); err != nil {
		t.Fatalf("unexpected submit error: %v", err)
	}

	p.Stop()

	snap := p.Snapshot()

	if snap.Stats.Processed != 1 {
		t.Fatalf("expected processed=1, got %d", snap.Stats.Processed)
	}

	if snap.QueueLength != 0 {
		t.Fatalf("expected queue length=0, got %d", snap.QueueLength)
	}
}

func TestProcessorRetriesFailedJob(t *testing.T) {
	p := NewProcessor(1, 10)

	if err := p.Start(); err != nil {
		t.Fatalf("unexpected start error: %v", err)
	}

	job := Job{
		ID: "job-retry",
		Transaction: &domain.Transaction{
			ID:        200,
			ToUserID:  1,
			Amount:    -10,
			Type:      domain.TransactionTypeCredit,
			Status:    domain.TransactionStatusPending,
			CreatedAt: time.Now(),
		},
		Ctx:        context.Background(),
		MaxRetries: 2,
	}

	if err := p.Submit(job); err != nil {
		t.Fatalf("unexpected submit error: %v", err)
	}

	time.Sleep(100 * time.Millisecond)
	p.Stop()

	if got := p.Stats().Retried(); got != 2 {
		t.Fatalf("expected retried=2, got %d", got)
	}

	if got := p.Stats().Failed(); got != 1 {
		t.Fatalf("expected failed=1, got %d", got)
	}
}
