// Package authorization defines request-level authorization data
// extracted by middleware and propagated via context.
package authorization

// Auth represents resolved access permissions for a request.
//
// The structure is populated by authorization middleware and stored in context.
// It does not perform authentication or permission checks itself.
type Auth struct {
	Read  bool
	Write bool
}
