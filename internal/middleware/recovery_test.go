package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecovery(t *testing.T) {

	//TODO  t.Run

	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	middleware := Recovery(log)

	tests := []struct {
		name        string
		writeHeader bool
		want        int
		handler     http.HandlerFunc
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
		handlerFunc := middleware(tt.handler)

		rr := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "http://example.com", http.NoBody)

		handlerFunc.ServeHTTP(rr, r)

		if rr.Code != tt.want {
			t.Errorf("%s got %d want %d", tt.name, rr.Code, tt.want)
		}
	}

}
