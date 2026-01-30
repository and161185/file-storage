package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"file-storage/internal/errs"
	"file-storage/internal/handlers/httpdto"
	"fmt"
	"net/http"
	"strings"
)

func validateUploadRequest(r *httpdto.UploadRequest) error {

	if err := validateUploadID(r.ID); err != nil {
		return err
	}

	//nothig useful in query
	if len(r.Data) == 0 {
		return errs.ErrNoDataToUpload
	}

	sum := sha256.Sum256(r.Data)
	hash := hex.EncodeToString(sum[:])
	if len(r.Data) != 0 && r.Hash != hash {
		return errs.ErrHashMismatch
	}

	if r.IsImage != nil && *r.IsImage {
		if !isImage(r.Data) {
			return errs.ErrNotSupportedImageType
		}
	}

	if r.Metadata != nil {
		for k, v := range r.Metadata {
			if err := checkMetadataValue(v); err != nil {
				return fmt.Errorf("field %s in metadata: %w", k, err)
			}
		}
	}

	return nil
}

func isImage(data []byte) bool {
	contentType := http.DetectContentType(data)
	return strings.HasPrefix(contentType, "image/")
}

func validateID(ID string) error {
	const IDLength = 36

	if len(ID) != IDLength {
		return fmt.Errorf("ID must be %d symbols: %w", IDLength, errs.ErrWrongIDLength)
	}

	return nil
}

func validateUploadID(ID string) error {
	const IDLength = 36

	if len(ID) != 0 && len(ID) != IDLength {
		return fmt.Errorf("ID must be empty or %d symbols: %w", IDLength, errs.ErrWrongIDLength)
	}

	return nil
}

func checkMetadataValue(v any) error {
	switch v.(type) {
	case string, bool, float64:
		return nil
	default:
		return errs.ErrUnsupportedTypeInMetadata
	}
}

func mapErrorToHttpStatus(err error) (int, bool) {
	switch {
	case errors.Is(err, errs.ErrHashMismatch),
		errors.Is(err, errs.ErrNoDataToUpload),
		errors.Is(err, errs.ErrInvalidImage):
		return http.StatusUnprocessableEntity, true

	case errors.Is(err, errs.ErrNotSupportedImageType),
		errors.Is(err, errs.ErrUnsupportedImageFormat):
		return http.StatusUnsupportedMediaType, true

	case errors.Is(err, errs.ErrWrongIDLength),
		errors.Is(err, errs.ErrMultipleIDsInQuery),
		errors.Is(err, errs.ErrWrongUrlParameter):
		return http.StatusBadRequest, true

	case errors.Is(err, errs.ErrNotFound):
		return http.StatusNotFound, true

	case errors.Is(err, errs.ErrAccessDenied):
		return http.StatusForbidden, true

	default:
		return http.StatusInternalServerError, false
	}
}
