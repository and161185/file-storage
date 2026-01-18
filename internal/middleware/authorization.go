package middleware

import (
	"context"
	"file-storage/internal/authorization"
	"file-storage/internal/config"
	"file-storage/internal/contextkeys"
	"net/http"
	"strings"
)

// Authorization middleware resolves access permissions and stores them in context.
func Authorization(security config.Security) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			token := r.Header.Get("Authorization")
			if token == "" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			tokenSlice := strings.Fields(token)
			if len(tokenSlice) < 2 ||
				strings.ToLower(tokenSlice[0]) != "bearer" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			tokenValue := tokenSlice[len(tokenSlice)-1]

			auth := authorization.Auth{}
			switch tokenValue {
			case security.ReadToken:
				auth.Read = true
				auth.Write = false
			case security.WriteToken:
				auth.Read = true
				auth.Write = true
			default:
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			ctxAuth := context.WithValue(r.Context(), contextkeys.ContextKeyAuth, &auth)
			r = r.WithContext(ctxAuth)

			next.ServeHTTP(w, r)
		})
	}
}
