package limiter

import (
	"sync"
	"time"
)

type Limiter struct {
	Capacity   float64
	RefillRate float64
	LastRefill time.Time
	Tokens     float64
	mx         sync.Mutex
}

func (l *Limiter) Allow() bool {
	if l.Capacity == 0 {
		return true
	}

	l.mx.Lock()
	defer l.mx.Unlock()

	now := time.Now()
	seconds := now.Sub(l.LastRefill).Seconds()
	l.Tokens = min(l.Tokens+l.RefillRate*seconds, l.Capacity)
	l.LastRefill = now

	if l.Tokens < 1 {
		return false
	}

	l.Tokens -= 1

	return true
}
