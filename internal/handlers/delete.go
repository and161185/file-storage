package handlers

import (
	"file-storage/internal/authorization"
	"file-storage/internal/contextkeys"
	"file-storage/internal/logger"
	"net/http"
	"strings"

	"github.com/go-chi/chi"
)

// Delete removes a file by ID.
// The operation is idempotent.
func DeleteHandler(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		log := logger.FromContext(ctx)
		log = logger.WithHandler(log, logger.HandlerDelete)

		auth, ok := ctx.Value(contextkeys.ContextKeyAuth).(*authorization.Auth)
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

		err = svc.Delete(ctx, ID)
		if err != nil {
			handleBusinessError(w, log, err)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
