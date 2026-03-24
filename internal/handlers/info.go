package handlers

import (
	"encoding/json"
	"file-storage/internal/authorization"
	"file-storage/internal/contextkeys"
	"file-storage/internal/errs"
	"file-storage/internal/logger"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

// Info returns file metadata by ID without file content.
func InfoHandler(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		log := logger.FromContext(ctx)
		log = logger.WithHandler(log, logger.HandlerInfo)

		auth, ok := ctx.Value(contextkeys.ContextKeyAuth).(*authorization.Auth)
		if !ok {
			err := fmt.Errorf("failed to get Auth structure out of context: %w", errs.ErrContextValueError)
			handleTransportError(w, log, err)
			return
		}
		if !auth.Read {
			err := fmt.Errorf("read access denied: %w", errs.ErrAccessDenied)
			handleTransportError(w, log, err)
			return
		}

		ID := strings.TrimSpace(chi.URLParam(r, "id"))

		err := validateID(ID)
		if err != nil {
			handleTransportError(w, log, err)
			return
		}

		fi, err := svc.Info(ctx, ID)
		if err != nil {
			handleBusinessError(w, log, err)
			return
		}

		body, err := json.Marshal(fi)
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
