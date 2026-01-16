// Package contextkeys defines typed keys for storing values in context.
//
// The package provides type-safe context keys to avoid collisions
// between different layers and packages.
package contextkeys

type contextKeyRequestID struct{}
type contextKeyAuth struct{}
type contextKeyLogger struct{}

var ContextKeyRequestID contextKeyRequestID = contextKeyRequestID{}
var ContextKeyAuth contextKeyAuth = contextKeyAuth{}
var ContextKeyLogger contextKeyLogger = contextKeyLogger{}
