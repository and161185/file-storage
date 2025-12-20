package middleware

import "net/http"

type responseWriter struct {
	http.ResponseWriter
	written    bool
	statusCode int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	if !w.written {
		w.written = true
		w.statusCode = statusCode
		w.ResponseWriter.WriteHeader(statusCode)
	}
}

func (w *responseWriter) Write(b []byte) (int, error) {
	if !w.written {
		w.written = true
		w.statusCode = http.StatusOK
	}

	return w.ResponseWriter.Write(b)
}
