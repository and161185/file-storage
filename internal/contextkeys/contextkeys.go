package contextkeys

type contextKeyRequestID struct{}
type contextKeyAuth struct{}

var ContextKeyRequestID contextKeyRequestID = contextKeyRequestID{}
var ContextKeyAuth contextKeyAuth = contextKeyAuth{}
