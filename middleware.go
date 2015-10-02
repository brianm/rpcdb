package drdb

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/mailgun/multibuf"
	"github.com/mailgun/oxy/utils"
)

// Middleware represents the middleware
type Middleware struct{}

// NewMiddleware creates the middleware
func NewMiddleware() *Middleware {
	return &Middleware{}
}

func isDebug(*http.Request) bool {
	return true
}

func (l *Middleware) ServeHTTP(w http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	body, err := multibuf.New(req.Body, multibuf.MaxBytes(128*1024*1024), multibuf.MemBytes(1024*1024))
	if err != nil || body == nil {
		panic(err)
	}
	totalSize, err := body.Size()
	if err != nil {
		panic(err)
	}

	debugreq := copyRequest(req, body, totalSize)
	newBody := debugRequest(debugreq)
	newMultiBuf, err := multibuf.New(newBody, multibuf.MaxBytes(128*1024*1024), multibuf.MemBytes(1024*1024))
	if _, err := body.Seek(0, 0); err != nil {
		panic(err)
	}

	outreq := copyRequest(req, newMultiBuf, totalSize)
	next(w, outreq)
	fmt.Println("after send")
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

func debugRequest(req *http.Request) io.Reader {
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
