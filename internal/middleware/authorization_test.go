package middleware

import (
	"file-storage/internal/authorization"
	"file-storage/internal/config"
	"file-storage/internal/contextkeys"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestAuthorization(t *testing.T) {

	var gotAuth any
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Context().Value(contextkeys.ContextKeyAuth)
	})

	s := config.Security{ReadToken: "111", WriteToken: "222"}
	middleware := Authorization(s)
	handlerFunc := middleware(handler)

	table := []struct {
		name       string
		auth       string
		wantStatus int
		wantAuth   *authorization.Auth
	}{
		{name: "unauthorized wrong", auth: "bearer 443", wantStatus: http.StatusUnauthorized},
		{name: "unauthorized no header", auth: "", wantStatus: http.StatusUnauthorized},
		{name: "authorized read", auth: "bearer 111", wantStatus: -1, wantAuth: &authorization.Auth{Read: true, Write: false}},
		{name: "authorized write", auth: "bearer 222", wantStatus: -1, wantAuth: &authorization.Auth{Read: true, Write: true}},
	}

	for _, tt := range table {
		w := httptest.NewRecorder()
		wwrapped := &responseWriter{w, false, -1}
		r := httptest.NewRequest("POST", "/method", http.NoBody)

		if tt.auth != "" {
			r.Header.Add("Authorization", tt.auth)
		}
		handlerFunc.ServeHTTP(wwrapped, r)

		if wwrapped.statusCode != tt.wantStatus {
			t.Errorf("test '%s': got %d want %d", tt.name, wwrapped.statusCode, tt.wantStatus)
		}

		if tt.wantAuth != nil {
			if !reflect.DeepEqual(gotAuth, tt.wantAuth) {
				t.Errorf("test '%s':  got %v want %v", tt.name, gotAuth, tt.wantAuth)
			}
		}
	}

}
