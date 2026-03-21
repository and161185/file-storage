package middleware

import (
	"file-storage/internal/limiter"
	"file-storage/internal/logger"
	"net/http"
)

func ConcurrencyLimiter(limiter *limiter.ConcurrencyLimiter) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Acquire() {
				w.WriteHeader(http.StatusServiceUnavailable)
				return
			}
			defer func() {
				err := limiter.Release()
				if err != nil {
					log := logger.FromContext(r.Context())
					log.Error("concurrency limiter release error", "error", err)
				}
			}()

			h.ServeHTTP(w, r)

		})
	}
}
