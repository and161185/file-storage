package contextkeys

type contextKeyRequestID struct{}
type contextKeyAuth struct{}
type contextKeyLogger struct{}

var ContextKeyRequestID contextKeyRequestID = contextKeyRequestID{}
var ContextKeyAuth contextKeyAuth = contextKeyAuth{}
var ContextKeyLogger contextKeyLogger = contextKeyLogger{}
