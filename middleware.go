package rpcdb

import "net/http"

// Middleware represents the middleware
type middleware struct {
	name string
	next http.Handler
}

// NewMiddleware directly builds the middleware handler
func NewMiddleware(name string, next http.Handler) http.Handler {
	return &middleware{name, next}
}

// Constructor returns a function that creates middleware for the
// given service name. This exists for Alice middleware chains.
func Constructor(name string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return &middleware{name, next}
	}
}

func (m *middleware) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if isDebug(req) {

	} else {
		m.next.ServeHTTP(w, req)
	}
}

func isDebug(req *http.Request) bool {
	if _, ok := req.Header[http.CanonicalHeaderKey("Debug-Breakpoint")]; ok {
		if _, ok := req.Header[http.CanonicalHeaderKey("Debug-Session")]; ok {
			return true
		}
	}
	return false
}
