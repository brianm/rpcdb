package rpcdb

import (
	"net/http"
)

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
		m.serveDebug(w, req)
	} else {
		m.next.ServeHTTP(w, req)
	}
}

func (m middleware) serveDebug(w http.ResponseWriter, req *http.Request) {
	session, err := BuildSession(m.name, req.Header)
	if err != nil {
		m.failWithError(w, err)
	}

	// receive hook
	debugRequest, err := session.Receive(req)
	if err != nil {
		m.failWithError(w, err)
		return
	}

	// reply hook
	// TODO consider StartReply to Reply which takes a closure. More ruby than go, but reliable.
	reply := session.StartReply(w, debugRequest)
	m.next.ServeHTTP(reply.CaptureWriter(), debugRequest)
	err = reply.FinishReply()
	if err != nil {
		m.failWithError(w, err)
	}
}

func isDebug(req *http.Request) bool {
	if _, ok := req.Header[debugBreakpointHeaderKey]; ok {
		if _, ok := req.Header[debugSessionHeaderKey]; ok {
			return true
		}
	}
	return false
}

func (m middleware) failWithError(w http.ResponseWriter, e error) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(500)
	w.Write([]byte(e.Error()))
}
