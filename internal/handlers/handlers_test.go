package handlers

import (
	"bytes"
	"context"
	"file-storage/internal/authorization"
	"file-storage/internal/contextkeys"
	"file-storage/internal/filedata"
	"file-storage/internal/logger"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi"
)

type mockService struct {
	fnUpdate  func(ctx context.Context, uc *filedata.UploadCommand) (string, error)
	fnContent func(ctx context.Context, cc *filedata.ContentCommand) ([]byte, error)
	fnInfo    func(ctx context.Context, ID string) (*filedata.FileInfo, error)
	fnDelete  func(ctx context.Context, ID string) error
}

func (s *mockService) Update(ctx context.Context, uc *filedata.UploadCommand) (string, error) {
	return s.fnUpdate(ctx, uc)
}
func (s *mockService) Content(ctx context.Context, cc *filedata.ContentCommand) ([]byte, error) {
	return s.fnContent(ctx, cc)
}
func (s *mockService) Info(ctx context.Context, ID string) (*filedata.FileInfo, error) {
	return s.fnInfo(ctx, ID)
}
func (s *mockService) Delete(ctx context.Context, ID string) error {
	return s.fnDelete(ctx, ID)
}

func newHttpTestRequest(method, target, body string) *http.Request {
	reader := bytes.NewReader([]byte(body))

	return httptest.NewRequest(method, target, reader)
}

func newContext(a *authorization.Auth, params map[string]string) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.ContextKeyLogger, logger.NewBootstrap())

	if a != nil {
		ctx = context.WithValue(ctx, contextkeys.ContextKeyAuth, a)
	}

	if params != nil {
		rctx := chi.NewRouteContext()
		for k, v := range params {
			rctx.URLParams.Add(k, v)
		}
		ctx = context.WithValue(ctx, chi.RouteCtxKey, rctx)
	}

	return ctx
}
