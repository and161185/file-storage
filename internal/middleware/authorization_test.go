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
	var handlerCalled bool
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Context().Value(contextkeys.ContextKeyAuth)
		handlerCalled = true
	})

	s := config.Security{ReadToken: "111", WriteToken: "222"}
	middleware := Authorization(s)
	handlerFunc := middleware(handler)

	table := []struct {
		name       string
		auth       string
		wantCalled bool
		wantStatus int
		wantAuth   *authorization.Auth
	}{
		{name: "unauthorized wrong", auth: "Bearer 443", wantCalled: false, wantStatus: http.StatusUnauthorized},
		{name: "unauthorized no header", auth: "", wantCalled: false, wantStatus: http.StatusUnauthorized},
		{name: "authorized read", auth: "Bearer 111", wantCalled: true, wantStatus: -1, wantAuth: &authorization.Auth{Read: true, Write: false}},
		{name: "authorized write", auth: "Bearer 222", wantCalled: true, wantStatus: -1, wantAuth: &authorization.Auth{Read: true, Write: true}},
	}

	for _, tt := range table {

		t.Run(tt.name, func(t *testing.T) {
			gotAuth = nil
			handlerCalled = false

			w := httptest.NewRecorder()
			wwrapped := &responseWriter{w, false, -1}
			r := httptest.NewRequest("POST", "/method", http.NoBody)

			if tt.auth != "" {
				r.Header.Add("Authorization", tt.auth)
			}
			handlerFunc.ServeHTTP(wwrapped, r)

			if wwrapped.statusCode != tt.wantStatus {
				t.Errorf("got %d want %d", wwrapped.statusCode, tt.wantStatus)
			}

			if handlerCalled != tt.wantCalled {
				t.Errorf("got %v want %v", handlerCalled, tt.wantCalled)
			}

			if tt.wantAuth != nil {
				if !reflect.DeepEqual(gotAuth, tt.wantAuth) {
					t.Errorf("got %v want %v", gotAuth, tt.wantAuth)
				}
			}
		})
	}

}
