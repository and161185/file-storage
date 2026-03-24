package handlers

import (
	"encoding/json"
	"file-storage/internal/authorization"
	"file-storage/internal/contextkeys"
	"file-storage/internal/errs"
	"file-storage/internal/filedata"
	"file-storage/internal/handlers/httpdto"
	"file-storage/internal/logger"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/google/uuid"
)

// Upload handles file upload or update requests.
func UploadHandler(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var ur httpdto.UploadRequest

		ctx := r.Context()
		log := logger.FromContext(ctx)
		log = logger.WithHandler(log, logger.HandlerUpdate)

		auth, ok := ctx.Value(contextkeys.ContextKeyAuth).(*authorization.Auth)
		if !ok {
			err := fmt.Errorf("failed to get Auth structure out of context: %w", errs.ErrContextValueError)
			handleTransportError(w, log, err)
			return
		}
		if !auth.Write {
			err := fmt.Errorf("write access denied: %w", errs.ErrAccessDenied)
			handleTransportError(w, log, err)
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
			handleTransportError(w, log, err)
			return
		}

		if ur.ID == "" {
			ur.ID = uuid.New().String()
		}

		var uc filedata.UploadCommand
		uc.ID = ur.ID
		uc.Hash = ur.Hash
		uc.Public = ur.Public
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
