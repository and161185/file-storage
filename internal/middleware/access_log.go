package middleware

import (
	"file-storage/internal/contextkeys"
	"file-storage/internal/logger"
	"net/http"
	"time"
)

func AccessLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log := logger.FromContext(r.Context())

		now := time.Now()
		next.ServeHTTP(w, r)

		statusCode := -1
		rw, ok := w.(*responseWriter)
		if ok {
			statusCode = rw.statusCode
		}

		contextRequestID := r.Context().Value(contextkeys.ContextKeyRequestID)
		if contextRequestID == nil {
			contextRequestID = ""
		}

		duration := time.Since(now).Microseconds()

		var fields []any
		fields = append(fields,
			logger.LogFieldMethod, r.Method,
			logger.LogFieldPath, r.URL.Path,
			logger.LogFieldStatus, statusCode,
			logger.LogFieldDuration, duration,
			logger.LogFieldRequestID, contextRequestID,
		)

		l := logger.WithComponent(log, logger.ComponentMiddleware)
		l = logger.WithMiddleware(l, logger.MiddlewareAccessLog)
		l.Info("request", fields...)
	})
}
