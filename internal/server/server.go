// Package server assembles and runs the file service application.
package server

import (
	"context"
	"file-storage/internal/config"
	"file-storage/internal/files"
	"file-storage/internal/handlers"
	"file-storage/internal/limiter"
	"file-storage/internal/middleware"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Server struct {
	service        *files.Service
	host           string
	port           int
	maxHeaderBytes int
	limits         Limits
	timeouts       Timeouts
	httpServer     *http.Server
	Log            *slog.Logger
}

type Limits struct {
	sizelimit          int
	RateLimiter        *limiter.RateLimiter
	concurrencyLimiter *limiter.ConcurrencyLimiter
}

type Timeouts struct {
	handlerTimeout    time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
}

func NewServer(config *config.App, svc *files.Service, log *slog.Logger) *Server {
	return &Server{
		host:           config.Server.Host,
		port:           config.Server.Port,
		maxHeaderBytes: config.Server.MaxHeaderBytes,
		limits: Limits{
			sizelimit:          config.Limits.SizeLimit,
			RateLimiter:        limiter.NewRateLimiter(&config.Limits.RateLimiter),
			concurrencyLimiter: limiter.NewConcurrencyLimiter(config.Limits.ConcurrencyLimit),
		},
		timeouts: Timeouts{
			handlerTimeout: config.Timeouts.HandlerTimeout,
		},
		service: svc,
		Log:     log,
	}
}

func (s *Server) Run(ctx context.Context, authCfg config.Security) error {

	r := chi.NewRouter()
	r.Use(middleware.RequestID(s.Log))
	r.Use(middleware.Recovery)
	r.Use(middleware.AccessLog)
	r.Use(middleware.Metrics)

	r.Group(func(r chi.Router) {
		r.Use(middleware.RateLimiter(s.limits.RateLimiter))
		r.Use(middleware.Timeout(s.timeouts.handlerTimeout))
		r.Use(middleware.SizeLimit(int64(s.limits.sizelimit)))
		r.Use(middleware.Authorization(authCfg))

		r.Get("/files/{id}/info", handlers.InfoHandler(s.service))
		r.Get("/files/{id}/content", handlers.ContentHandler(s.service))
		r.Post("/files/upload", handlers.UploadHandler(s.service))
		r.Delete("/files/{id}/delete", handlers.DeleteHandler(s.service))
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.Timeout(s.timeouts.handlerTimeout))
		r.Get("/files/metrics", promhttp.Handler().ServeHTTP)
	})

	s.httpServer = &http.Server{
		Addr:              ":" + strconv.Itoa(s.port),
		BaseContext:       func(l net.Listener) context.Context { return ctx },
		Handler:           r,
		MaxHeaderBytes:    s.maxHeaderBytes,
		WriteTimeout:      s.timeouts.writeTimeout,
		ReadHeaderTimeout: s.timeouts.readHeaderTimeout,
		IdleTimeout:       s.timeouts.idleTimeout,
	}

	listener, err := net.Listen("tcp", net.JoinHostPort(s.host, strconv.Itoa(s.port)))
	if err != nil {
		return fmt.Errorf("listen error: %w", err)
	}
	s.Log.Info("server listening")
	err = s.httpServer.Serve(listener)
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("serve error: %w", err)
	}

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}

	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		return err
	}
	return nil
}
