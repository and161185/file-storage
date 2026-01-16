// Package logger provides logger constructors and helper functions.
package logger

import (
	"context"
	"file-storage/internal/config"
	"file-storage/internal/contextkeys"
	"log/slog"
	"os"
	"strings"
)

type ComponentName string
type MiddlewareName string
type HandlerName string

const (
	ComponentMiddleware ComponentName = "middleware"
)

const (
	MiddlewareRecovery  MiddlewareName = "recovery"
	MiddlewareAccessLog MiddlewareName = "access_log"
)

const (
	HandlerContent HandlerName = "content"
	HandlerDelete  HandlerName = "delete"
	HandlerInfo    HandlerName = "info"
	HandlerUpdate  HandlerName = "upload"
)

const (
	LogFieldRequestID = "request_id"
	LogFieldMethod    = "method"
	LogFieldPath      = "path"
	LogFieldStatus    = "status"
	LogFieldDuration  = "duration"
	LogFieldPanic     = "panic"
	LogFieldStack     = "stack"
	LogFieldError     = "error"
)

func NewBootstrap() *slog.Logger {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil)).With("phase", "bootstrap")
	return logger
}

func New(cfg *config.Log) *slog.Logger {

	level := slog.LevelInfo
	switch strings.ToLower(cfg.Level) {
	case config.LogLevelDebug:
		level = slog.LevelDebug
	case config.LogLevelError:
		level = slog.LevelError
	case config.LogLevelWarn:
		level = slog.LevelWarn
	}
	opts := &slog.HandlerOptions{Level: level}

	var logger *slog.Logger
	if cfg.Type == config.LogTypeJSON {
		logger = slog.New(slog.NewJSONHandler(os.Stdout, opts))
	} else {
		logger = slog.New(slog.NewTextHandler(os.Stdout, opts))
	}
	return logger
}

func WithComponent(log *slog.Logger, c ComponentName) *slog.Logger {
	return log.With("component", c)
}

func WithMiddleware(log *slog.Logger, m MiddlewareName) *slog.Logger {
	return log.With("middleware", m)
}

func WithHandler(log *slog.Logger, h HandlerName) *slog.Logger {
	return log.With("handler", h)
}

func FromContext(ctx context.Context) *slog.Logger {
	l, ok := ctx.Value(contextkeys.ContextKeyLogger).(*slog.Logger)

	if !ok || l == nil {
		panic("logger is missing in context; RequestID middleware put it into context and must run first")
	}
	return l
}
