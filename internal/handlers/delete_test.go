package handlers

import (
	"context"
	"file-storage/internal/authorization"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteHandler(t *testing.T) {

	correctID := "012345678901234567890123456789012345"

	table := []struct {
		name       string
		service    *mockService
		ctx        context.Context
		request    *http.Request
		wantStatus int
	}{
		{
			name:       "no auth structure in context",
			service:    &mockService{},
			ctx:        newContext(nil, nil),
			request:    newHttpTestRequest("DELETE", "/", ""),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "no rights",
			service:    &mockService{},
			ctx:        newContext(&authorization.Auth{}, nil),
			request:    newHttpTestRequest("DELETE", "/", ""),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "invalid ID",
			service:    &mockService{},
			ctx:        newContext(&authorization.Auth{Write: true}, map[string]string{"id": "21"}),
			request:    newHttpTestRequest("DELETE", "/", ""),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "business error",
			service: &mockService{fnDelete: func(ctx context.Context, ID string) error {
				return fmt.Errorf("error")
			}},
			ctx:        newContext(&authorization.Auth{Write: true}, map[string]string{"id": correctID}),
			request:    newHttpTestRequest("DELETE", "/", ""),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name: "ok",
			service: &mockService{fnDelete: func(ctx context.Context, ID string) error {
				return nil
			}},
			ctx:        newContext(&authorization.Auth{Write: true}, map[string]string{"id": correctID}),
			request:    newHttpTestRequest("DELETE", "/", ""),
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			handlerFunc := DeleteHandler(tt.service)

			w := httptest.NewRecorder()
			r := tt.request
			r = r.WithContext(tt.ctx)
			handlerFunc.ServeHTTP(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %v want %v", w.Code, tt.wantStatus)
			}
		})
	}
}
