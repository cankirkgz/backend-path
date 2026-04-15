package processor

import (
	"sync"
)

type Processor struct {
	mu          sync.Mutex
	workerCount int
	queue       chan Job
	wg          sync.WaitGroup
	stats       *Stats
	started     bool
	stopped     bool
}

func NewProcessor(workerCount, queueSize int) *Processor {
	return &Processor{
		workerCount: workerCount,
		queue:       make(chan Job, queueSize),
		stats:       &Stats{},
	}
}

func (p *Processor) Start() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.started {
		return ErrProcessorAlreadyStarted
	}

	if p.stopped {
		return ErrProcessorStopped
	}

	p.started = true

	for i := 0; i < p.workerCount; i++ {
		p.wg.Add(1)

		go func(workerID int) {
			defer p.wg.Done()

			for job := range p.queue {
				p.handleJob(workerID, job)
			}
		}(i + 1)
	}

	return nil
}

func (p *Processor) Submit(job Job) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.stopped {
		return ErrProcessorStopped
	}

	p.queue <- job
	return nil
}

func (p *Processor) Stop() {
	p.mu.Lock()

	if p.stopped {
		p.mu.Unlock()
		return
	}

	p.stopped = true
	close(p.queue)
	p.mu.Unlock()

	p.wg.Wait()
}

func (p *Processor) shouldRetry(job Job) bool {
	return job.RetryCount < job.MaxRetries
}

func (p *Processor) handleJob(workerID int, job Job) {
	_ = workerID

	if job.Ctx != nil {
		select {
		case <-job.Ctx.Done():
			p.stats.IncCancelled()
			return
		default:
		}
	}

	if job.Transaction == nil {
		p.stats.IncFailed()
		return
	}

	if err := job.Transaction.Validate(); err != nil {
		if p.shouldRetry(job) {
			job.RetryCount++
			p.stats.IncRetried()
			_ = p.Submit(job)
			return
		}

		_ = job.Transaction.MarkFailed()
		p.stats.IncFailed()
		return
	}

	if err := job.Transaction.MarkCompleted(); err != nil {
		if p.shouldRetry(job) {
			job.RetryCount++
			p.stats.IncRetried()
			_ = p.Submit(job)
			return
		}

		p.stats.IncFailed()
		return
	}

	p.stats.IncProcessed()
}

func (p *Processor) Stats() *Stats {
	return p.stats
}

func (p *Processor) SubmitBatch(jobs []Job) error {
	for _, job := range jobs {
		if err := p.Submit(job); err != nil {
			return err
		}
	}
	return nil
}

func (p *Processor) QueueLength() int {
	return len(p.queue)
}

type ProcessorSnapshot struct {
	QueueLength int
	Stats       StatsSnapshot
}

func (p *Processor) Snapshot() ProcessorSnapshot {
	return ProcessorSnapshot{
		QueueLength: p.QueueLength(),
		Stats:       p.stats.Snapshot(),
	}
}
