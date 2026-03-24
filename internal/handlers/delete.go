package handlers

import (
	"file-storage/internal/authorization"
	"file-storage/internal/contextkeys"
	"file-storage/internal/errs"
	"file-storage/internal/logger"
	"fmt"
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
			err := fmt.Errorf("failed to get Auth structure out of context: %w", errs.ErrContextValueError)
			handleTransportError(w, log, err)
			return
		}
		if !auth.Write {
			err := fmt.Errorf("write access denied: %w", errs.ErrAccessDenied)
			handleTransportError(w, log, err)
			return
		}

		ID := strings.TrimSpace(chi.URLParam(r, "id"))

		err := validateID(ID)
		if err != nil {
			handleTransportError(w, log, err)
			return
		}

		err = svc.Delete(ctx, ID)
		if err != nil {
			handleBusinessError(w, log, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
