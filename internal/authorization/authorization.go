// Package authorization defines authorization models used across the service.
//
// The package contains transport-level authorization data extracted from
// authentication middleware and propagated via context.
package authorization

// Auth represents resolved access permissions for a request.
//
// The structure is populated by authorization middleware and stored in context.
// It does not perform authentication or permission checks itself.
type Auth struct {
	Read  bool
	Write bool
}
