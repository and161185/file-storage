package middleware

import (
	"context"
	"net/http"
	"time"
)

func Timeout(sec int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			ctx, cancel := context.WithTimeout(r.Context(), time.Duration(sec)*time.Second)
			defer cancel()

			rt := r.WithContext(ctx)

			next.ServeHTTP(w, rt)
		})
	}
}
