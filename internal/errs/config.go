package errs

import "errors"

var ErrConfigFlagsNotParsed = errors.New("flags not parsed")
var ErrConfigPortOutOfRange = errors.New("config app port out of range")
var ErrConfigWrongLogLevel = errors.New("wrong config log level. should be debug or info or warn or error")
var ErrConfigWrongLogType = errors.New("wrong config log type. should be json or text")
