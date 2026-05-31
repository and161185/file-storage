// Package authorization defines resolved request access flags
// propagated through context by authorization middleware.
package authorization

// Auth represents resolved read/write access flags for a request.
//
// The structure is populated by authorization middleware and stored in context.
// It does not contain user identity, subject, claims or token metadata.
// It does not perform authentication or permission checks itself.
type Auth struct {
	Read  bool
	Write bool
}
