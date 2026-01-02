package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"file-storage/internal/errs"
	"file-storage/internal/logger"
	"file-storage/internal/models"
	"log/slog"
	"net/http"
)

func UploadHandler(log *slog.Logger) http.HandlerFunc {

	return func(w http.ResponseWriter, r *http.Request) {
		var fd models.UploadRequest

		log := logger.FromContext(r.Context())
		log = logger.WithHandler(log, logger.HandlerUpdate)

		decoder := json.NewDecoder(r.Body)
		decoder.DisallowUnknownFields()
		err := decoder.Decode(&fd)
		if err != nil {
			http.Error(w, "invalid request payload", http.StatusBadRequest)
			log.Error("failed to read body", slog.Any(logger.LogFieldError, err))
			return
		}

		err = ValidateUploadRequest(&fd)
		if err != nil {
			handleValidationError(w, log, err)
			return
		}
	}
}

func ValidateUploadRequest(r *models.UploadRequest) error {

	//nothig useful in query
	if len(r.Data) == 0 && len(r.Metadata) == 0 {
		return errs.ErrNoDataToUpload
	}

	//if no data the only thing we can do is update metadata
	if len(r.Data) == 0 && r.ID == "" {
		return errs.ErrMissingIdToUpdateMetadata
	}

	sum := sha256.Sum256(r.Data)
	hash := hex.EncodeToString(sum[:])
	if len(r.Data) != 0 && r.Hash != hash {
		return errs.ErrHashMismatch
	}

	return nil
}

func handleValidationError(w http.ResponseWriter, log *slog.Logger, err error) {

	log.Warn("payload validation failed", slog.Any(logger.LogFieldError, err))

	status, handledError := mapErrorToHttpStatus(err)
	if !handledError {
		log.Warn("unhandled error", slog.Any(logger.LogFieldError, err))
	}

	http.Error(w, err.Error(), status)

}

func mapErrorToHttpStatus(err error) (int, bool) {
	switch {
	case errors.Is(err, errs.ErrHashMismatch):
		return http.StatusUnprocessableEntity, true
	case errors.Is(err, errs.ErrNoDataToUpload):
		return http.StatusUnprocessableEntity, true
	case errors.Is(err, errs.ErrMissingIdToUpdateMetadata):
		return http.StatusUnprocessableEntity, true
	default:
		return http.StatusInternalServerError, false
	}
}
