package limiter

import (
	"file-storage/internal/errs"
	"sync"
)

type ConcurrencyLimiter struct {
	capacity int
	inFlight int
	mx       sync.Mutex
}

func NewConcurrencyLimiter(limit int) *ConcurrencyLimiter {
	return &ConcurrencyLimiter{
		capacity: limit,
		inFlight: 0,
	}
}

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
