package handlers

import (
	"context"
	"file-storage/internal/authorization"
	"file-storage/internal/errs"
	"file-storage/internal/filedata"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestContentHandler(t *testing.T) {

	correctID := "012345678901234567890123456789012345"

	table := []struct {
		name       string
		service    *mockService
		ctx        context.Context
		request    *http.Request
		wantStatus int
	}{
		{
			name:       "invalid id",
			service:    &mockService{},
			ctx:        newContext(&authorization.Auth{Read: true}, map[string]string{"id": "12"}),
			request:    newHttpTestRequest("GET", "/", ""),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid width",
			service:    &mockService{},
			ctx:        newContext(&authorization.Auth{Read: true}, map[string]string{"id": "12"}),
			request:    newHttpTestRequest("GET", "/method?width=err", ""),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid height",
			service:    &mockService{},
			ctx:        newContext(&authorization.Auth{Read: true}, map[string]string{"id": "12"}),
			request:    newHttpTestRequest("GET", "/method?height=err", ""),
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid format",
			service: &mockService{fnContent: func(ctx context.Context, cc *filedata.ContentCommand) ([]byte, error) {
				return nil, errs.ErrUnsupportedImageFormat
			}},
			ctx:        newContext(&authorization.Auth{Read: true}, map[string]string{"id": correctID}),
			request:    newHttpTestRequest("GET", "/method?format=err", ""),
			wantStatus: http.StatusUnsupportedMediaType,
		},
		{
			name: "ok",
			service: &mockService{fnContent: func(ctx context.Context, cc *filedata.ContentCommand) ([]byte, error) {
				return []byte("ok"), nil
			}},
			ctx:        newContext(&authorization.Auth{Read: true}, map[string]string{"id": correctID}),
			request:    newHttpTestRequest("GET", "/method", ""),
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			handlerFunc := ContentHandler(tt.service)

			w := httptest.NewRecorder()
			r := tt.request.WithContext(tt.ctx)
			handlerFunc.ServeHTTP(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d want %d; responce %s", w.Code, tt.wantStatus, w.Body)
			}
		})
	}
}
