package limiter

import (
	"errors"
	"file-storage/internal/errs"
	"testing"
)

func TestAcquire(t *testing.T) {
	table := []struct {
		name string
		l    *ConcurrencyLimiter
		want []bool
	}{
		{
			name: "out of cap",
			l:    NewConcurrencyLimiter(1),
			want: []bool{true, false},
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			for _, want := range tt.want {
				result := tt.l.Acquire()
				if result != want {
					t.Errorf("got %v want %v", result, want)
				}
			}
		})
	}
}

func TestRelease(t *testing.T) {

	table := []struct {
		name string
		l    *ConcurrencyLimiter
		want []error
	}{
		{
			name: "ok",
			l:    NewConcurrencyLimiter(1),
			want: []error{nil},
		},
		{
			name: "ok no limiter",
			l:    NewConcurrencyLimiter(0),
			want: []error{nil, nil},
		},
		{
			name: "release error",
			l:    NewConcurrencyLimiter(1),
			want: []error{nil, errs.ErrConcurrencyLimiterBelow0},
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {

			tt.l.Acquire()

			for _, want := range tt.want {
				err := tt.l.Release()
				if !errors.Is(err, want) {
					t.Errorf("got %v want %v", err, want)
				}
			}
		})
	}
}
