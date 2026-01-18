package handlers

import (
	"encoding/json"
	"file-storage/internal/authorization"
	"file-storage/internal/contextkeys"
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
		log = logger.WithHandler(log, logger.HandlerDelete)

		auth, ok := ctx.Value(contextkeys.ContextKeyAuth).(authorization.Auth)
		if !ok {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("failed to get Auth structure out of context")
			return
		}
		if !auth.Read {
			w.WriteHeader(http.StatusForbidden)
			log.Warn("read access denied")
			return
		}

		ID := strings.TrimSpace(chi.URLParam(r, "id"))

		err := validateID(ID)
		if err != nil {
			handleValidationError(w, log, err)
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
