package errs

import "errors"

var ErrConfigFlagsNotParsed = errors.New("flags not parsed")
var ErrConfigPortOutOfRange = errors.New("config app port out of range")
var ErrConfigWrongLogLevel = errors.New("invalid config log level. should be debug or info or warn or error")
var ErrConfigWrongLogType = errors.New("invalid config log type. should be json or text")
var ErrConfigInvalidImageFormat = errors.New("invalid image format. Only bmp, jpg, png, gif, webp are supported")
var ErrConfigImageDimentionOutOfRange = errors.New("Stored image dimention out of range 1000 - 10000")
var ErrTokenNotSet = errors.New("token not set")
