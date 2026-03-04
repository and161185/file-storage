package middleware

import (
	"file-storage/internal/metrics"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
)

func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()

		metrics.HTTPrequestsInFlight.Inc()
		defer metrics.HTTPrequestsInFlight.Dec()

		next.ServeHTTP(w, r)

		operation := chi.RouteContext(r.Context()).RoutePattern()
		if operation == "" {
			operation = "unknown"
		}

		metricsQuery := strings.Contains(operation, "/metrics")

		if metricsQuery {
			return
		}

		statusCode := -1
		rw, ok := w.(*responseWriter)
		if ok {
			statusCode = rw.statusCode
		}

		metrics.HTTPrequestsTotal.WithLabelValues(operation, strconv.Itoa(statusCode)).Inc()
		metrics.HTTPrequestsDurationSeconds.WithLabelValues(operation).Observe(time.Since(start).Seconds())
	})
}
