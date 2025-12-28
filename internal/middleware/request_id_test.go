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
		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/method", http.NoBody)

		if tt.id != "" {
			r.Header.Add(HeaderRequestIDName, tt.id)
		}

		handlerFunc.ServeHTTP(w, r)

		grID, ok := gotRequestID.(string)
		if !ok {
			t.Errorf("test '%s': type assertion failed", tt.name)
			continue
		}
		if tt.id != "" {
			if tt.id != grID {
				t.Errorf("test '%s':  got %s want %s", tt.name, grID, tt.id)
			}
		} else {
			if grID == "" {
				t.Errorf("test '%s':  got %s want new ID", tt.name, grID)
			}
		}

	}

}
