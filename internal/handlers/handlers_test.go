package handlers

import (
	"bytes"
	"context"
	"file-storage/internal/authorization"
	"file-storage/internal/contextkeys"
	"file-storage/internal/files"
	"file-storage/internal/logger"
	"net/http"
	"net/http/httptest"
)

type mockService struct {
	fnUpdate  func(ctx context.Context, uc *files.UploadCommand) (string, error)
	fnContent func(ctx context.Context, cc *files.ContentCommand) ([]byte, error)
	fnInfo    func(ctx context.Context, ID string) (*files.FileInfo, error)
	fnDelete  func(ctx context.Context, ID string) error
}

func (s *mockService) Update(ctx context.Context, uc *files.UploadCommand) (string, error) {
	return s.fnUpdate(ctx, uc)
}
func (s *mockService) Content(ctx context.Context, cc *files.ContentCommand) ([]byte, error) {
	return s.fnContent(ctx, cc)
}
func (s *mockService) Info(ctx context.Context, ID string) (*files.FileInfo, error) {
	return s.fnInfo(ctx, ID)
}
func (s *mockService) Delete(ctx context.Context, ID string) error {
	return s.fnDelete(ctx, ID)
}

func newHttpTestRequest(method, target, body string) *http.Request {
	reader := bytes.NewReader([]byte(body))

	return httptest.NewRequest(method, target, reader)
}

func newContext(a *authorization.Auth) context.Context {
	ctx := context.Background()
	ctx = context.WithValue(ctx, contextkeys.ContextKeyLogger, logger.NewBootstrap())

	if a == nil {
		return ctx
	}

	return context.WithValue(ctx, contextkeys.ContextKeyAuth, *a)
}
