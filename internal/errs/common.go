package errs

import "errors"

var ErrNotFound = errors.New("not found")
var ErrWrongFormat = errors.New("format unsupported")
var ErrTooBig = errors.New("too big")
