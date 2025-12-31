package middleware

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTimeout(t *testing.T) {

	ch := make(chan error, 1)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		<-ctx.Done()

		ch <- ctx.Err()
	})

	middleware := Timeout(20)
	handlerFunc := middleware(handler)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/method", http.NoBody)

	handlerFunc.ServeHTTP(w, r)

	errCtx := <-ch
	if !errors.Is(errCtx, context.DeadlineExceeded) {
		t.Errorf("got %s want %s", errCtx, context.DeadlineExceeded)
	}

}
