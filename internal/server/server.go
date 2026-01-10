package server

import (
	"context"
	"file-storage/internal/config"
	"file-storage/internal/files"
	"file-storage/internal/handlers"
	"log/slog"
	"net"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type Server struct {
	Service *files.Service
	port    int
	Log     *slog.Logger
}

func NewServer(config *config.App, svc *files.Service, log *slog.Logger) *Server {
	return &Server{port: config.Port, Service: svc, Log: log}
}

func (s *Server) Run(ctx context.Context) error {

	r := chi.NewRouter()
	//r.Use(middleware.Logger)
	//r.Get("/", handlers.Get)
	r.Get("/files/{id}/info", handlers.InfoHandler(s.Service))
	r.Get("/files/{id}/content", handlers.ContentHandler(s.Service))
	r.Post("/files/upload", handlers.UploadHandler(s.Service))
	r.Delete("/files/{id}", handlers.DeleteHandler(s.Service))
	//httpServer.ListenAndServe(":"+strconv.Itoa(s.port), r)
	httpServer := http.Server{Addr: ":" + strconv.Itoa(s.port),
		BaseContext: func(l net.Listener) context.Context { return ctx },
		Handler:     r}
	httpServer.ListenAndServe()

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {

	return nil
}
