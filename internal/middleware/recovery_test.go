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

func TestRecovery(t *testing.T) {

	tests := []struct {
		name    string
		want    int
		handler http.HandlerFunc
	}{
		{
			name: "panic",
			want: http.StatusInternalServerError,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				panic("panic")
			}),
		},
		{
			name: "panic with header written",
			want: http.StatusOK,
			handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				panic("panic")
			}),
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {

			handlerFunc := Recovery(tt.handler)

			rr := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "http://example.com", http.NoBody)
			log := slog.New(slog.NewTextHandler(io.Discard, nil))
			ctx := context.WithValue(r.Context(), contextkeys.ContextKeyLogger, log)
			r = r.WithContext(ctx)

			handlerFunc.ServeHTTP(rr, r)

			if rr.Code != tt.want {
				t.Errorf("got %d want %d", rr.Code, tt.want)
			}
		})

	}

}
