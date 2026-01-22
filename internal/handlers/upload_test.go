package handlers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"file-storage/internal/authorization"
	"file-storage/internal/errs"
	"file-storage/internal/filedata"
	"file-storage/internal/handlers/httpdto"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUploadHandler(t *testing.T) {

	ur := httpdto.UploadRequest{Data: []byte("123")}
	bodyInvalidHash, err := json.Marshal(ur)
	if err != nil {
		t.Fatalf("upload request preparation fail: %v", err)
	}

	ur = httpdto.UploadRequest{Data: []byte("123")}
	sum := sha256.Sum256(ur.Data)
	ur.Hash = hex.EncodeToString(sum[:])
	bodyOK, err := json.Marshal(ur)
	if err != nil {
		t.Fatalf("upload request preparation fail: %v", err)
	}

	table := []struct {
		name       string
		service    *mockService
		ctx        context.Context
		request    *http.Request
		wantStatus int
	}{
		{
			name:       "no auth structure",
			service:    &mockService{},
			ctx:        newContext(nil, nil),
			request:    newHttpTestRequest("POST", "/", ""),
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "no rights",
			service:    &mockService{},
			ctx:        newContext(&authorization.Auth{Write: false}, nil),
			request:    newHttpTestRequest("POST", "/", ""),
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "no body",
			service:    &mockService{},
			ctx:        newContext(&authorization.Auth{Write: true}, nil),
			request:    newHttpTestRequest("POST", "/", ""),
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid body",
			service:    &mockService{},
			ctx:        newContext(&authorization.Auth{Write: true}, nil),
			request:    newHttpTestRequest("POST", "/", "{}"),
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "invalid hash",
			service:    &mockService{},
			ctx:        newContext(&authorization.Auth{Write: true}, nil),
			request:    newHttpTestRequest("POST", "/", string(bodyInvalidHash)),
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name: "unsupported image",
			service: &mockService{fnUpdate: func(ctx context.Context, uc *filedata.UploadCommand) (string, error) {
				return "", errs.ErrUnsupportedImageFormat
			}},
			ctx:        newContext(&authorization.Auth{Write: true}, nil),
			request:    newHttpTestRequest("POST", "/", string(bodyOK)),
			wantStatus: http.StatusUnsupportedMediaType,
		},
		{
			name: "ok",
			service: &mockService{fnUpdate: func(ctx context.Context, uc *filedata.UploadCommand) (string, error) {
				return "", nil
			}},
			ctx:        newContext(&authorization.Auth{Write: true}, nil),
			request:    newHttpTestRequest("POST", "/", string(bodyOK)),
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range table {
		t.Run(tt.name, func(t *testing.T) {
			handler := UploadHandler(tt.service)

			r := tt.request.WithContext(tt.ctx)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d want %d; responce %s", w.Code, tt.wantStatus, w.Body)
			}
		})
	}

}
