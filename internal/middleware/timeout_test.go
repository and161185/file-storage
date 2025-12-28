package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTimeout(t *testing.T) {

	var errCtx error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		select {
		case <-ctx.Done():
		}

		errCtx = ctx.Err()
	})

	middleware := Timeout(1)
	handlerFunc := middleware(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/method", http.NoBody)
	handlerFunc.ServeHTTP(w, r)

	if errCtx != context.DeadlineExceeded {
		t.Errorf("got %s want %s", errCtx, context.DeadlineExceeded)
	}

}
