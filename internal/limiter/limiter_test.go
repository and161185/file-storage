package limiter

import (
	"testing"
	"time"
)

func TestAllow(t *testing.T) {
	table := []struct {
		name         string
		l            *Limiter
		iters_expect []bool
		timeout      time.Duration
	}{
		{
			name:         "no limiter",
			l:            &Limiter{Capacity: 0, RefillRate: 10, LastRefill: time.Now(), Tokens: 0},
			iters_expect: []bool{true},
		},
		{
			name:         "no tokens",
			l:            &Limiter{Capacity: 1, RefillRate: 10, LastRefill: time.Now(), Tokens: 0},
			iters_expect: []bool{false},
		},
		{
			name:         "tokens out",
			l:            &Limiter{Capacity: 1, RefillRate: 10, LastRefill: time.Now(), Tokens: 1},
			iters_expect: []bool{true, false},
		},
		{
			name:         "clamp to capacity",
			l:            &Limiter{Capacity: 1, RefillRate: 0, LastRefill: time.Now(), Tokens: 22},
			iters_expect: []bool{true, false},
		},
		{
			name:         "ok",
			l:            &Limiter{Capacity: 2, RefillRate: 0, LastRefill: time.Now(), Tokens: 2},
			iters_expect: []bool{true, true},
		},
		{
			name:         "ok refill",
			l:            &Limiter{Capacity: 2, RefillRate: 100, LastRefill: time.Now(), Tokens: 1},
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
