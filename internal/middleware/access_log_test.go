package middleware

import (
	"context"
	"file-storage/internal/contextkeys"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccessLog(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	handlerFunc := AccessLog(handler)

	rec := httptest.NewRecorder()
	wwrapped := &responseWriter{rec, false, -1}

	req := httptest.NewRequest("POST", "/ololo", http.NoBody)
	ctx := context.WithValue(req.Context(), contextkeys.ContextKeyLogger, log)
	req = req.WithContext(ctx)

	handlerFunc.ServeHTTP(wwrapped, req)

	if wwrapped.statusCode != 200 {
		t.Errorf("want 200 get %d", wwrapped.statusCode)
	}

}
