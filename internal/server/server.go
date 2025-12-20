package server

import (
	"context"
	"file-storage/internal/config"
	"file-storage/internal/handlers"
	"log/slog"
	"net"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
)

type Server struct {
	port int
	Log  *slog.Logger
}

func NewServer(config *config.App, log *slog.Logger) *Server {
	return &Server{port: config.Port, Log: log}
}

func (s *Server) Run(ctx context.Context) error {

	r := chi.NewRouter()
	//r.Use(middleware.Logger)
	r.Get("/", handlers.Get)
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
