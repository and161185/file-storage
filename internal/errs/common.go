package errs

import "errors"

var ErrNotFound = errors.New("not found")
var ErrWrongFormat = errors.New("format unsupported")
var ErrTooBig = errors.New("data is too big")
var ErrHashMismatch = errors.New("provided hash doesn’t match calculated one")
var ErrNoDataToUpload = errors.New("no data to upload")
var ErrMissingIdToUpdateMetadata = errors.New("missing id to update metadata")
