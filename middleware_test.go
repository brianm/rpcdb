package rpcdb

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStuffHappens(t *testing.T) {

	handler := Stub{}

	m := NewMiddleware(handler)
	w := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "http://example.com/hello", strings.NewReader("hello"))
	req.Header.Add("Debug-Session", "http://127.0.0.1:1234/abc123")
	req.Header.Add("Debug-Breakpoint", "receive example:/hello") // example server receives /hello
	req.Header.Add("Debug-Breakpoint", "reply example:/hello")   // example server responds to /hello
	req.Header.Add("Debug-Breakpoint", "request example:*")      // example server issues any rpc

	if err != nil {
		log.Fatal(err)
	}

	m.ServeHTTP(w, req)
}

type Stub struct{}

func (s Stub) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("hello world"))
}
