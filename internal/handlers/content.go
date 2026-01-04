package handlers

import (
	"file-storage/internal/authorization"
	"file-storage/internal/contextkeys"
	"file-storage/internal/errs"
	"file-storage/internal/handlers/models"
	"file-storage/internal/logger"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi"
)

func ContentHandler(w http.ResponseWriter, r *http.Request) {
	log := logger.FromContext(r.Context())
	log = logger.WithHandler(log, logger.HandlerContent)

	auth, ok := r.Context().Value(contextkeys.ContextKeyAuth).(authorization.Auth)
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

	contentRequest, err := parseContentRequest(r)
	if err != nil {
		handleValidationError(w, log, err)
		return
	}

	contentRequest.ID = ID

	//TODO business logic content
}

func parseContentRequest(r *http.Request) (*models.ContentRequest, error) {
	contentRequest := models.ContentRequest{}

	q := r.URL.Query()

	widthParam := q.Get("width")
	if widthParam != "" {
		width, err := strconv.Atoi(widthParam)
		if err != nil {
			return nil, fmt.Errorf("invalid width param %q: %w", widthParam, errs.ErrWrongUrlParameter)
		}
		contentRequest.Width = &width
	}

	heightParam := q.Get("height")
	if heightParam != "" {
		height, err := strconv.Atoi(heightParam)
		if err != nil {
			return nil, fmt.Errorf("invalid height param %q: %w", heightParam, errs.ErrWrongUrlParameter)
		}
		contentRequest.Height = &height
	}

	format := q.Get("format")
	if format != "" {
		contentRequest.Format = &format
	}

	return &contentRequest, nil
}
