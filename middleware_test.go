package rpcdb

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"strings"
	"testing"

	"github.com/justinas/alice"
)

func ExampleConstructor() {
	chain := alice.New(Constructor("example")).ThenFunc(handler)

	err := http.ListenAndServe("127.0.0.1:3000", chain)
	if err != nil {
		log.Panicf("unable to start: %s", err)
	}
}

func handler(w http.ResponseWriter, req *http.Request) {
	bytes, err := httputil.DumpRequest(req, true)
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "%s", err)
		return
	}
	w.Write(bytes)
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

func Test500OnBadBreakpoint(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("debug-breakpoint", "buggy example:*")
	req.Header.Add("debug-session", "http://example/123")

	handler := Stub{200, []byte("hello world")}
	m := NewMiddleware("example", handler)
	w := httptest.NewRecorder()

	m.ServeHTTP(w, req)
	if w.Code != 500 {
		t.Errorf("Expected 500 status code, got %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "text/plain" {
		body, _ := ioutil.ReadAll(w.Body)
		t.Log(string(body))
		t.Errorf("Expected text/plain body, got %s", w.Header().Get("Content-Type"))
	}

}

// TODO convert test to use a stub which checks the body rather than relies
//      on echoing the output back.z
func TestReceiveBodyTransform(t *testing.T) {
	// stand up a mocked rpcdb daemon
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		// always respond with a body replacement
		fmt.Fprintln(w, `{"body":"howdy world"}`)
	}))
	defer ts.Close()

	// nil body makes the stub return whatever the input was :-)
	handler := Stub{200, nil}
	m := NewMiddleware("example", handler)
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "http://example.com/hello", strings.NewReader("hello world"))

	req.Header.Add("Debug-Session", ts.URL)                      // debug session URL
	req.Header.Add("Debug-Breakpoint", "receive example:/hello") // server receives /hello

	m.ServeHTTP(w, req)

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("error reading response body: %s", err)
	}
	if string(body) != "howdy world" {
		t.Errorf("Expected body to be transformed to 'howdy world' got '%s'", body)
	}
}

func TestReplyBodyTransform(t *testing.T) {
	// stand up a mocked rpcdb daemon
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		// always respond with a body replacement
		fmt.Fprintln(w, `{"body":"TRANSFORMED"}`)
	}))
	defer ts.Close()

	// hardcode output as "hello world"
	handler := Stub{200, []byte("hello world")}
	m := NewMiddleware("example", handler)
	w := httptest.NewRecorder()

	req, _ := http.NewRequest("POST", "http://example.com/hello", strings.NewReader("ignored"))

	req.Header.Add("Debug-Session", ts.URL)                    // debug session URL
	req.Header.Add("Debug-Breakpoint", "reply example:/hello") // server receives /hello

	m.ServeHTTP(w, req)

	body, err := ioutil.ReadAll(w.Body)
	if err != nil {
		t.Fatalf("error reading response body: %s", err)
	}
	if string(body) != "TRANSFORMED" {
		t.Errorf("Expected body to be transformed to 'TRANSFORMED' got '%s'", body)
	}

}

type Stub struct {
	code int
	body []byte
}

func (s Stub) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(s.code)
	if s.body == nil {
		bytes, _ := ioutil.ReadAll(req.Body)
		req.Body.Close()
		w.Write(bytes)
	} else {
		w.Write(s.body)
	}
}
