package handlers

import (
	"file-storage/internal/errs"
	"file-storage/internal/filedata"
	"file-storage/internal/handlers/httpdto"
	"file-storage/internal/logger"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

// Content returns file content by ID.
// The handler returns raw bytes and does not include metadata.
// doesn't need authorisation
func ContentHandler(svc Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		log := logger.FromContext(ctx)
		log = logger.WithHandler(log, logger.HandlerContent)

		ID := strings.TrimSpace(chi.URLParam(r, "id"))

		cr, err := parseContentRequest(r, ID)
		if err != nil {
			handleValidationError(w, log, err)
			return
		}

		cc := filedata.ContentCommand{
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

func parseContentRequest(r *http.Request, ID string) (*httpdto.ContentRequest, error) {
	contentRequest := httpdto.ContentRequest{}

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
