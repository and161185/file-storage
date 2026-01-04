package handlers

import (
	"encoding/json"
	"file-storage/internal/authorization"
	"file-storage/internal/contextkeys"
	"file-storage/internal/handlers/models"
	"file-storage/internal/logger"
	"log/slog"
	"net/http"
	"strings"
)

func UploadHandler(w http.ResponseWriter, r *http.Request) {
	var fd models.UploadRequest

	log := logger.FromContext(r.Context())
	log = logger.WithHandler(log, logger.HandlerUpdate)

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

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	decoder.DisallowUnknownFields()
	err := decoder.Decode(&fd)
	if err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		log.Error("failed to read body", slog.Any(logger.LogFieldError, err))
		return
	}

	fd.ID = strings.TrimSpace(fd.ID)
	err = validateUploadRequest(&fd)
	if err != nil {
		handleValidationError(w, log, err)
		return
	}

	// TODO: call upload business layer
}

func handleValidationError(w http.ResponseWriter, log *slog.Logger, err error) {

	log.Warn("query validation failed", slog.Any(logger.LogFieldError, err))

	status, handledError := mapErrorToHttpStatus(err)
	if !handledError {
		log.Warn("unhandled error", slog.Any(logger.LogFieldError, err))
	}

	http.Error(w, err.Error(), status)

}
