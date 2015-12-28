package rpcdb

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func _TestExample(t *testing.T) {
	// stand up an rpcdb daemon at http://<something or other>
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		fmt.Fprintln(w, `{"body":"howdy world"}`)
	}))
	defer ts.Close()

	// start a debug session named "abc123"

	handler := Stub{200, []byte("hello world")}
	m := NewMiddleware("example", handler)
	w := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "http://example.com/hello", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("unable to create test request: %s", err)
	}

	req.Header.Add("Debug-Session", ts.URL)                      // debug session URL
	req.Header.Add("Debug-Breakpoint", "receive example:/hello") // example server receives /hello
	//req.Header.Add("Debug-Breakpoint", "reply example:/hello")   // example server responds to /hello
	//req.Header.Add("Debug-Breakpoint", "request example:*")      // example server issues any rpc
	//req.Header.Add("Debug-Breakpoint", "response example:*")     // example server gets response to any rpc
	//req.Header.Add("Debug-Breakpoint", "receive other:*")        // other server receives any, will be proxied along

	m.ServeHTTP(w, req)

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("error reading response body: %s", err)
	}
	if string(body) != "howdy" {
		t.Errorf("Expected body to be transformed to 'howdy world' got '%s'", body)
	}
	// rpcdbd will be called for 'receive example:/hello'
	// tell rpcdbd to proceed, but change body to "howdy"
	// stub is invoked, body is "howdy"
	// rpcdbd will be called for 'reply example:/hello' with body "hello world"
	// tell rpcdbd to proceed, but change body to "howdy world!"
	// check that w.Body.String() == "howdy world!"
	// ^^ all assumes semaphores in stub and so on to do the right things :-)
}

func TestIsDebugCanonicalHeaders(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("Debug-Breakpoint", "receive example:*")
	req.Header.Add("Debug-Session", "http://example/123")

	if !isDebug(req) {
		t.Error("expected request to be debug=true, was not")
	}
}

func TestIsDebugNonCanonicalHeaders(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("debug-breakpoint", "receive example:*")
	req.Header.Add("debug-session", "http://example/123")

	if !isDebug(req) {
		t.Error("expected request to be debug=true, was not")
	}
}

func _TestReceiveTransform(t *testing.T) {
	// stand up an rpcdb daemon at http://<something or other>
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		fmt.Fprintln(w, `{"body":"howdy world"}`)
	}))
	defer ts.Close()

	// start a debug session named "abc123"

	handler := Stub{200, []byte("hello world")}
	m := NewMiddleware("example", handler)
	w := httptest.NewRecorder()

	req, err := http.NewRequest("POST", "http://example.com/hello", strings.NewReader("hello"))
	if err != nil {
		t.Fatalf("unable to create test request: %s", err)
	}

	req.Header.Add("Debug-Session", ts.URL)                      // debug session URL
	req.Header.Add("Debug-Breakpoint", "receive example:/hello") // example server receives /hello

	m.ServeHTTP(w, req)

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("error reading response body: %s", err)
	}
	if string(body) != "howdy" {
		t.Errorf("Expected body to be transformed to 'howdy world' got '%s'", body)
	}
}

type Stub struct {
	code int
	body []byte
}

func (s Stub) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(s.code)
	w.Write(s.body)
}
