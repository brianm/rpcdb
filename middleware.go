package rpcdb

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/mailgun/multibuf"
	"github.com/mailgun/oxy/utils"
)

// Middleware represents the middleware
type Middleware struct {
	next http.Handler
}

// NewMiddleware creates the middleware
func NewMiddleware(next http.Handler) http.Handler {
	return &Middleware{next: next}
}

func isDebug(*http.Request) bool {
	return true
}

type debugResponseWriter struct {
	header http.Header
}

// Header returns the header map that will be sent by
// WriteHeader. Changing the header after a call to
// WriteHeader (or Write) has no effect unless the modified
// headers were declared as trailers by setting the
// "Trailer" header before the call to WriteHeader (see example).
// To suppress implicit response headers, set their value to nil.
// Header() Header
func (w *debugResponseWriter) Header() http.Header {
	return w.header
}

// Write writes the data to the connection as part of an HTTP reply.
// If WriteHeader has not yet been called, Write calls WriteHeader(http.StatusOK)
// before writing the data.  If the Header does not contain a
// Content-Type line, Write adds a Content-Type set to the result of passing
// the initial 512 bytes of written data to DetectContentType.
// Write([]byte) (int, error)
func (w *debugResponseWriter) Write(buf []byte) (int, error) {
	panic("not implemented yet")
}

// WriteHeader sends an HTTP response header with status code.
// If WriteHeader is not called explicitly, the first call to Write
// will trigger an implicit WriteHeader(http.StatusOK).
// Thus explicit calls to WriteHeader are mainly used to
// send error codes.
func (w *debugResponseWriter) WriteHeader(status int) {
	panic("not implemented yet")
}

func copyReponseWriter(w http.ResponseWriter) *debugResponseWriter {
	drw := &debugResponseWriter{
		header: make(map[string][]string),
	}
	// copy header values from w?
	log.Println("copy response doesn't actually copy anything yet!")
	return drw
}

// TODO convert to alice style middleware
func (l *Middleware) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := multibuf.New(req.Body, multibuf.MaxBytes(128*1024*1024), multibuf.MemBytes(1024*1024))
	if err != nil || body == nil {
		panic(err)
	}
	totalSize, err := body.Size()
	if err != nil {
		panic(err)
	}

	// trigger DEBUG RECEIVE hook
	// right now we just handle body, but we should probably allow new
	// headers as well as the new body.
	debugreq := copyRequest(req, body, totalSize)
	newBody := debugReceive(debugreq)
	newMultiBuf, err := multibuf.New(newBody, multibuf.MaxBytes(128*1024*1024), multibuf.MemBytes(1024*1024))
	if _, err := body.Seek(0, 0); err != nil {
		panic(err)
	}

	// now issue request to the server proper
	outreq := copyRequest(req, newMultiBuf, totalSize)
	l.next.ServeHTTP(w, outreq)

	// now need trigger RPC REPLY hook
	// need to make an http.ResponseWriter which will buffer the response and
	// allow us to hit the hook and possibly respond with a different value.
}

// from mailgun/oxy streamer
func copyRequest(req *http.Request, body io.ReadCloser, bodySize int64) *http.Request {
	o := *req
	o.URL = utils.CopyURL(req.URL)
	o.Header = make(http.Header)
	utils.CopyHeaders(o.Header, req.Header)
	o.ContentLength = bodySize

	// remove TransferEncoding that could have been previously set because we
	// have transformed the request from chunked encoding
	o.TransferEncoding = []string{}

	// http.Transport will close the request body on any error, we are controlling
	// the close process ourselves, so we override the closer here
	o.Body = ioutil.NopCloser(body)
	return &o
}

// RPC RECEIVE debug hook
func debugReceive(req *http.Request) io.Reader {
	buf, err := ioutil.ReadAll(req.Body)
	if err != nil {
		panic(err)
	}
	if strings.Contains(string(buf), "joe") {
		return bytes.NewBuffer([]byte(`{"name":"bob"}`))
	}
	return bytes.NewBuffer(buf)
}

func debugResponse() {

}
