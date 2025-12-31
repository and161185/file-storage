package middleware

import (
	"file-storage/internal/contextkeys"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID(t *testing.T) {

	var gotRequestID any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotRequestID = r.Context().Value(contextkeys.ContextKeyRequestID)
	})

	handlerFunc := RequestID(handler)

	table := []struct {
		name string
		id   string
	}{
		{name: "test with id", id: "111-222"},
		{name: "test without id", id: ""},
	}

	for _, tt := range table {

		t.Run(tt.name, func(t *testing.T) {
			gotRequestID = nil

			w := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/method", http.NoBody)

			if tt.id != "" {
				r.Header.Add(HeaderRequestIDName, tt.id)
			}

			handlerFunc.ServeHTTP(w, r)

			grID, ok := gotRequestID.(string)
			if !ok {
				t.Fatalf("type assertion failed")
			}
			if tt.id != "" {
				if tt.id != grID {
					t.Errorf("got context request ID %s want %s", grID, tt.id)
				}
			} else {
				if grID == "" {
					t.Error("got context request ID as empty string want new ID")
				}
			}

			if tt.id != "" {
				if tt.id != w.Header().Get(HeaderRequestIDName) {
					t.Errorf("got emty header %s, want filled", HeaderRequestIDName)
				}
			} else {
				if w.Header().Get(HeaderRequestIDName) == "" {
					t.Errorf("got emty header %s, want filled", HeaderRequestIDName)
				}
			}
		})
	}
}
