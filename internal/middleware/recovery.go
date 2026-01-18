package middleware

import (
	"file-storage/internal/contextkeys"
	"file-storage/internal/logger"
	"net/http"
	"runtime/debug"
)

// Recover middleware catches panics and returns internal server error.
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		wwrapped := &responseWriter{w, false, 0}

		defer func() {
			if v := recover(); v != nil {

				var fields []any
				requestID := r.Context().Value(contextkeys.ContextKeyRequestID)

				if requestID != nil {
					fields = append(fields, logger.LogFieldRequestID, requestID)
				}
				fields = append(fields, logger.LogFieldMethod, r.Method,
					logger.LogFieldPath, r.URL.Path,
					logger.LogFieldPanic, v,
					logger.LogFieldStack, debug.Stack(),
				)

				l := logger.FromContext(r.Context())
				l = logger.WithComponent(l, logger.ComponentMiddleware)
				l = logger.WithMiddleware(l, logger.MiddlewareRecovery)
				l.Error("panic recovered", fields...)

				if !wwrapped.written {
					wwrapped.WriteHeader(http.StatusInternalServerError)
				}
			}
		}()

		next.ServeHTTP(wwrapped, r)

	})
}
