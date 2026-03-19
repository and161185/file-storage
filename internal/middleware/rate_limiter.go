package middleware

import (
	"file-storage/internal/limiter"
	"net/http"
)

func RateLimiter(l *limiter.Limiter) func(http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if l.Allow() {
				h.ServeHTTP(w, r)
			} else {
				w.WriteHeader(http.StatusTooManyRequests)
			}
		})
	}
}
