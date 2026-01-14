package handlers

import (
	"encoding/json"
	"file-storage/internal/authorization"
	"file-storage/internal/contextkeys"
	"file-storage/internal/files"
	"file-storage/internal/handlers/models"
	"file-storage/internal/logger"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
)

func UploadHandler(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ur models.UploadRequest

		ctx := r.Context()
		log := logger.FromContext(ctx)
		log = logger.WithHandler(log, logger.HandlerUpdate)

		auth, ok := ctx.Value(contextkeys.ContextKeyAuth).(authorization.Auth)
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
		err := decoder.Decode(&ur)
		if err != nil {
			http.Error(w, "invalid request payload", http.StatusBadRequest)
			log.Error("failed to read body", slog.Any(logger.LogFieldError, err))
			return
		}

		ur.ID = strings.TrimSpace(ur.ID)
		err = validateUploadRequest(&ur)
		if err != nil {
			handleValidationError(w, log, err)
			return
		}

		var uc files.UploadCommand
		uc.ID = ur.ID
		uc.Hash = ur.Hash
		uc.Data = ur.Data
		uc.Metadata = ur.Metadata
		if ur.IsImage == nil {
			isImage := isImage(uc.Data)
			uc.IsImage = isImage
		} else {
			uc.IsImage = *ur.IsImage
		}

		ID, err := svc.Update(ctx, &uc)
		if err != nil {
			handleBusinessError(w, log, err)
			return
		}

		body, err := json.Marshal(map[string]string{"id": ID})
		if err != nil {
			handleBusinessError(w, log, fmt.Errorf("marshalling error: %w", err))
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err = w.Write([]byte(body))
		if err != nil {
			log.Error("write body error", slog.Any(logger.LogFieldError, err))
		}

	}
}

func handleValidationError(w http.ResponseWriter, log *slog.Logger, err error) {

	log.Warn("query validation failed", slog.Any(logger.LogFieldError, err))

	status, handledError := mapErrorToHttpStatus(err)
	if !handledError {
		log.Warn("unhandled error", slog.Any(logger.LogFieldError, err))
	}

	writeErrorResponse(w, err, status)

}

func handleBusinessError(w http.ResponseWriter, log *slog.Logger, err error) {

	log.Error("execution error", slog.Any(logger.LogFieldError, err))

	status, handledError := mapErrorToHttpStatus(err)
	if !handledError {
		log.Error("unhandled internal error", slog.Any(logger.LogFieldError, err))
	}

	writeErrorResponse(w, err, status)

}

func writeErrorResponse(w http.ResponseWriter, err error, status int) {

	if status >= 500 {
		http.Error(w, http.StatusText(status), status)
	} else {
		http.Error(w, err.Error(), status)
	}

}
