package middleware

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSizeLimit(t *testing.T) {

	handlerCalled := false
	var handlerErr error
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true

		_, handlerErr = io.ReadAll(r.Body)
	})

	middleware := SizeLimit(8)
	handlerFunc := middleware(handler)

	table := []struct {
		name              string
		contentLength     string
		body              string
		wantHandlerCalled bool
		wantHandlerErr    error
		wantStatus        int
	}{
		{
			name:              "content length header too big",
			contentLength:     "15",
			body:              "",
			wantHandlerCalled: false,
			wantHandlerErr:    nil,
			wantStatus:        http.StatusRequestEntityTooLarge,
		},
		{
			name:              "content too big",
			contentLength:     "1",
			body:              "test string",
			wantHandlerCalled: true,
			wantHandlerErr:    &http.MaxBytesError{Limit: 8},
			wantStatus:        http.StatusOK,
		},
	}

	for _, tt := range table {
		w := httptest.NewRecorder()
		wwrapped := &responseWriter{w, false, 0}

		r := httptest.NewRequest("GET", "/method", strings.NewReader(tt.body))
		r.Header.Set("Content-Length", tt.contentLength)

		handlerCalled = false
		handlerErr = nil
		handlerFunc.ServeHTTP(wwrapped, r)

		if handlerCalled != tt.wantHandlerCalled {
			t.Errorf("test %s handler called got %v want %v", tt.name, handlerCalled, tt.wantHandlerCalled)
			continue
		}
		if tt.wantHandlerErr != nil {
			if handlerErr.Error() != tt.wantHandlerErr.Error() {
				t.Errorf("test %s handler err got %s want %s", tt.name, handlerErr, tt.wantHandlerErr)
				continue
			}
		}
		if w.Code != tt.wantStatus {
			t.Errorf("test %s status got %d want %d", tt.name, w.Code, tt.wantStatus)
			continue
		}
	}

}
