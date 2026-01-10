package handlers

import (
	"file-storage/internal/authorization"
	"file-storage/internal/contextkeys"
	"file-storage/internal/errs"
	"file-storage/internal/files"
	"file-storage/internal/handlers/models"
	"file-storage/internal/logger"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

func ContentHandler(svc *files.Service) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		log := logger.FromContext(ctx)
		log = logger.WithHandler(log, logger.HandlerContent)

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

		cr, err := parseContentRequest(r, ID)
		if err != nil {
			handleValidationError(w, log, err)
			return
		}

		cc := files.ContentCommand{
			ID:     cr.ID,
			Width:  cr.Width,
			Height: cr.Height,
			Format: cr.Format,
		}

		content, err := svc.Content(ctx, &cc)
		if err != nil {
			handleBusinessError(w, log, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(content)
	}
}

func parseContentRequest(r *http.Request, ID string) (*models.ContentRequest, error) {
	contentRequest := models.ContentRequest{}

	err := validateID(ID)
	if err != nil {
		return nil, err
	}
	contentRequest.ID = ID

	q := r.URL.Query()

	widthParam := strings.TrimSpace(q.Get("width"))
	if widthParam != "" {
		width, err := strconv.Atoi(widthParam)
		if err != nil || width < 10 {
			return nil, fmt.Errorf("invalid width param %q: %w", widthParam, errs.ErrWrongUrlParameter)
		}
		contentRequest.Width = &width
	}

	heightParam := strings.TrimSpace(q.Get("height"))
	if heightParam != "" {
		height, err := strconv.Atoi(heightParam)
		if err != nil || height < 10 {
			return nil, fmt.Errorf("invalid height param %q: %w", heightParam, errs.ErrWrongUrlParameter)
		}
		contentRequest.Height = &height
	}

	format := strings.TrimSpace(q.Get("format"))
	if format != "" {
		contentRequest.Format = &format
	}

	return &contentRequest, nil
}
