package rpcdb

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestExample(t *testing.T) {

	// stand up an rpcdb daemon at http://127.0.0.1:1234/
	// start a debug session named "abc123"

	handler := Stub{}
	m := NewMiddleware("example", handler)
	w := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "http://example.com/hello", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("unable to create test request: %s", err)
	}

	req.Header.Add("Debug-Session", "http://127.0.0.1:1234/session/abc123") // debug session URL
	req.Header.Add("Debug-Breakpoint", "receive example:/hello")            // example server receives /hello
	req.Header.Add("Debug-Breakpoint", "reply example:/hello")              // example server responds to /hello
	req.Header.Add("Debug-Breakpoint", "request example:*")                 // example server issues any rpc

	req.Header.Add("Debug-Breakpoint", "receive other:*") // other server receives any, will be proxied along

	go m.ServeHTTP(w, req)

	// rpcdbd will be called for 'receive example:/hello'
	// tell rpcdbd to proceed, but change body to "howdy"
	// stub is invoked, body is "howdy"
	// rpcdbd will be called for 'reply example:/hello' with body "hello world"
	// tell rpcdbd to proceed, but change body to "howdy world!"
	// check that w.Body.String() == "howdy world!"
	// ^^ all assumes semaphores in stub and so on to do the right things :-)
}

type Stub struct{}

func (s Stub) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("hello world"))
}
