package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"file-storage/internal/errs"
	"file-storage/internal/imgproc"
	"file-storage/internal/models"
	"fmt"
	"net/http"
)

func validateUploadRequest(r *models.UploadRequest) error {

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

	if r.IsImage != nil && *r.IsImage {
		if !imgproc.IsImage(r.Data) {
			return errs.ErrNotSupportedImageType
		}
	}

	if r.IsImage == nil {
		isImage := imgproc.IsImage(r.Data)
		r.IsImage = &isImage
	}

	return nil
}

func validateQueryID(ID string) error {
	const IDLength = 36

	if len(ID) != IDLength {
		return fmt.Errorf("ID must be %d symbols: %w", IDLength, errs.ErrWrongIDLength)
	}

	return nil
}

func mapErrorToHttpStatus(err error) (int, bool) {
	switch {
	case errors.Is(err, errs.ErrHashMismatch),
		errors.Is(err, errs.ErrNoDataToUpload),
		errors.Is(err, errs.ErrMissingIdToUpdateMetadata):
		return http.StatusUnprocessableEntity, true
	case errors.Is(err, errs.ErrNotSupportedImageType):
		return http.StatusUnsupportedMediaType, true
	case errors.Is(err, errs.ErrWrongIDLength),
		errors.Is(err, errs.ErrMultipleIDsInQuery):
		return http.StatusBadRequest, true
	default:
		return http.StatusInternalServerError, false
	}
}
