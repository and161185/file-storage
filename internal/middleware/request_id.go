package middleware

import (
	"context"
	"file-storage/internal/contextkeys"
	"net/http"

	"github.com/google/uuid"
)

const HeaderRequestIDName = "X-Request-ID"

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		rID := r.Header.Get(HeaderRequestIDName)
		if rID == "" {
			rID = uuid.New().String()
		}

		contextID := context.WithValue(r.Context(), contextkeys.ContextKeyRequestID, rID)
		r = r.WithContext(contextID)

		w.Header().Add(HeaderRequestIDName, rID)

		next.ServeHTTP(w, r)
	})
}
