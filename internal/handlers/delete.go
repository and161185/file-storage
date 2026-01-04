package handlers

import (
	"file-storage/internal/authorization"
	"file-storage/internal/contextkeys"
	"file-storage/internal/logger"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

func DeleteHandler(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log = logger.WithHandler(log, logger.HandlerDelete)

	auth, ok := r.Context().Value(contextkeys.ContextKeyAuth).(authorization.Auth)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		log.Error("failed to get Auth structure out of context")
		return
	}
	if !auth.Write {
		w.WriteHeader(http.StatusForbidden)
		log.Warn("write access denied")
		return
	}

	ID := strings.TrimSpace(chi.URLParam(r, "id"))

	err := validateID(ID)
	if err != nil {
		handleValidationError(w, log, err)
		return
	}

	//ID := IDs[0]
	//TODO business logic deletion
}
