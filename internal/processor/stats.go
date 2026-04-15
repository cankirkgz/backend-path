package processor

import "sync/atomic"

type Stats struct {
	processed uint64
	failed    uint64
	cancelled uint64
	retried   uint64
}

func (s *Stats) IncProcessed() {
	atomic.AddUint64(&s.processed, 1)
}

func (s *Stats) IncFailed() {
	atomic.AddUint64(&s.failed, 1)
}

func (s *Stats) IncCancelled() {
	atomic.AddUint64(&s.cancelled, 1)
}

func (s *Stats) Processed() uint64 {
	return atomic.LoadUint64(&s.processed)
}

func (s *Stats) Failed() uint64 {
	return atomic.LoadUint64(&s.failed)
}

func (s *Stats) Cancelled() uint64 {
	return atomic.LoadUint64(&s.cancelled)
}

type StatsSnapshot struct {
	Processed uint64
	Failed    uint64
	Cancelled uint64
	Retried   uint64
}

func (s *Stats) Snapshot() StatsSnapshot {
	return StatsSnapshot{
		Processed: atomic.LoadUint64(&s.processed),
		Failed:    atomic.LoadUint64(&s.failed),
		Cancelled: atomic.LoadUint64(&s.cancelled),
		Retried:   atomic.LoadUint64(&s.retried),
	}
}

func (s *Stats) IncRetried() {
	atomic.AddUint64(&s.retried, 1)
}

func (s *Stats) Retried() uint64 {
	return atomic.LoadUint64(&s.retried)
}
