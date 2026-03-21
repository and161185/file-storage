package limiter

import (
	"file-storage/internal/config"
	"sync"
	"time"
)

type RateLimiter struct {
	capacity   float64
	refillRate float64
	lastRefill time.Time
	tokens     float64
	mx         sync.Mutex
}

func NewRateLimiter(c *config.RateLimiter) *RateLimiter {
	return &RateLimiter{
		capacity:   float64(c.Capacity),
		refillRate: float64(c.RefillRate),
		lastRefill: time.Now(),
		tokens:     float64(c.Capacity),
	}
}

func (l *RateLimiter) Allow() bool {
	if l.capacity <= 0 {
		return true
	}

	l.mx.Lock()
	defer l.mx.Unlock()

	now := time.Now()
	seconds := now.Sub(l.lastRefill).Seconds()
	l.tokens = min(l.tokens+l.refillRate*seconds, l.capacity)
	l.lastRefill = now

	if l.tokens < 1 {
		return false
	}

	l.tokens -= 1

	return true
}
