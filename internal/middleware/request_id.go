package middleware

import (
	"context"
	"file-storage/internal/contextkeys"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
)

const HeaderRequestIDName = "X-Request-ID"

func RequestID(log *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			rID := r.Header.Get(HeaderRequestIDName)
			if rID == "" {
				rID = uuid.New().String()
			}

			l := log.With("request_id", rID)

			ctx := r.Context()
			ctx = context.WithValue(ctx, contextkeys.ContextKeyRequestID, rID)
			ctx = context.WithValue(ctx, contextkeys.ContextKeyLogger, l)
			r = r.WithContext(ctx)

			w.Header().Set(HeaderRequestIDName, rID)

			next.ServeHTTP(w, r)
		})
	}
}
