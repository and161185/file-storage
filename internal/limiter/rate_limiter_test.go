package limiter

import (
	"testing"
	"time"
)

func TestAllow(t *testing.T) {
	table := []struct {
		name         string
		l            *RateLimiter
		iters_expect []bool
		timeout      time.Duration
	}{
		{
			name:         "no limiter",
			l:            &RateLimiter{capacity: 0, refillRate: 10, lastRefill: time.Now(), tokens: 0},
			iters_expect: []bool{true},
		},
		{
			name:         "no tokens",
			l:            &RateLimiter{capacity: 1, refillRate: 10, lastRefill: time.Now(), tokens: 0},
			iters_expect: []bool{false},
		},
		{
			name:         "tokens out",
			l:            &RateLimiter{capacity: 1, refillRate: 10, lastRefill: time.Now(), tokens: 1},
			iters_expect: []bool{true, false},
		},
		{
			name:         "clamp to capacity",
			l:            &RateLimiter{capacity: 1, refillRate: 0, lastRefill: time.Now(), tokens: 22},
			iters_expect: []bool{true, false},
		},
		{
			name:         "ok",
			l:            &RateLimiter{capacity: 2, refillRate: 0, lastRefill: time.Now(), tokens: 2},
			iters_expect: []bool{true, true},
		},
		{
			name:         "ok refill",
			l:            &RateLimiter{capacity: 2, refillRate: 100, lastRefill: time.Now(), tokens: 1},
			iters_expect: []bool{true, true},
			timeout:      10 * time.Millisecond,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {

			for _, expect := range tt.iters_expect {
				result := tt.l.Allow()
				if tt.timeout > 0 {
					time.Sleep(tt.timeout)
				}
				if result != expect {
					t.Errorf("expect %v got %v", expect, result)
				}
			}
		})
	}
}
