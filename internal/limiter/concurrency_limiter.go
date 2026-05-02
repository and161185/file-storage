package limiter

import (
	"file-storage/internal/errs"
	"sync"
)

// ConcurrencyLimiter limits the number of requests that may be processed at the same time.
type ConcurrencyLimiter struct {
	capacity int
	inFlight int
	mx       sync.Mutex
}

// NewConcurrencyLimiter creates a concurrency limiter with the configured number of execution slots.
func NewConcurrencyLimiter(limit int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		capacity: limit,
		inFlight: 0,
	}
}

// Acquire attempts to reserve one execution slot and reports whether the request may proceed.
func (l *ConcurrencyLimiter) Acquire() bool {
	if l.capacity <= 0 {
		return true
	}

	l.mx.Lock()
	defer l.mx.Unlock()

	if l.inFlight >= l.capacity {
		return false
	}

	l.inFlight++

	return true
}

// Release returns one previously acquired execution slot back to the limiter.
func (l *ConcurrencyLimiter) Release() error {
	if l.capacity <= 0 {
		return nil
	}

	l.mx.Lock()
	defer l.mx.Unlock()

	l.inFlight--

	if l.inFlight < 0 {
		l.inFlight = 0
		return errs.ErrConcurrencyLimiterBelow0
	}

	return nil
}
