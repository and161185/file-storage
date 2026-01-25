// Package server assembles and runs the file service application.
package server

import (
	"context"
	"file-storage/internal/config"
	"file-storage/internal/files"
	"file-storage/internal/handlers"
	"file-storage/internal/middleware"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
)

type Server struct {
	Service    *files.Service
	host       string
	port       int
	sizelimit  int
	timeout    time.Duration
	httpServer *http.Server
	Log        *slog.Logger
}

func NewServer(config *config.App, svc *files.Service, log *slog.Logger) *Server {
	return &Server{port: config.Port, sizelimit: config.SizeLimit, timeout: config.Timeout, Service: svc, Log: log}
}

func (s *Server) Run(ctx context.Context, authCfg config.Security) error {

	r := chi.NewRouter()
	r.Use(middleware.RequestID(s.Log))
	r.Use(middleware.Recovery)
	r.Use(middleware.AccessLog)
	r.Use(middleware.Timeout(s.timeout))
	r.Use(middleware.SizeLimit(int64(s.sizelimit)))
	r.Use(middleware.Authorization(authCfg))

	r.Get("/files/{id}/info", handlers.InfoHandler(s.Service))
	r.Get("/files/{id}/content", handlers.ContentHandler(s.Service))
	r.Post("/files/upload", handlers.UploadHandler(s.Service))
	r.Delete("/files/{id}/delete", handlers.DeleteHandler(s.Service))

	s.httpServer = &http.Server{Addr: ":" + strconv.Itoa(s.port),
		BaseContext: func(l net.Listener) context.Context { return ctx },
		Handler:     r}

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
