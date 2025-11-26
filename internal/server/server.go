package server

import (
	"context"
	"file-storage/internal/config"
	"net/http"

	"github.com/go-chi/chi"
)

type Server struct {
}

func NewServer(config *config.Config) *Server {
	return &Server{}
}

func (s *Server) Run(ctx context.Context) error {
	r := chi.NewRouter()
	//r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("welcome"))
	})
	http.ListenAndServe(":3000", r)

	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {

	return nil
}
