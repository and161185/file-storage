package logger

import (
	"file-storage/internal/config"
	"log/slog"
	"os"
	"strings"
)

type ComponentName string
type MiddlewareName string

const (
	ComponentMiddleware ComponentName = "middleware"
)

const (
	MiddlewareRecovery  MiddlewareName = "recovery"
	MiddlewareAccessLog MiddlewareName = "access_log"
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
