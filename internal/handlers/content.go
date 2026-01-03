package handlers

import (
	"file-storage/internal/authorization"
	"file-storage/internal/contextkeys"
	"file-storage/internal/logger"
	"net/http"

	"github.com/go-chi/chi"
)

func ContentHandler(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log = logger.WithHandler(log, logger.HandlerContent)

	auth, ok := r.Context().Value(contextkeys.ContextKeyAuth).(authorization.Auth)
	if !ok {
		w.WriteHeader(http.StatusForbidden)
		log.Error("failed to get Auth structure out of context")
		return
	}
	if !auth.Read {
		w.WriteHeader(http.StatusForbidden)
		log.Warn("write access denied")
		return
	}

	ID := chi.URLParam(r, "id")

	err := validateQueryID(ID)
	if err != nil {
		handleValidationError(w, log, err)
		return
	}

	//ID := IDs[0]
	//TODO business logic deletion
}
