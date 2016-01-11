package rpcdb

import (
	"fmt"
	"github.com/alioygur/gores"
	"golang.org/x/net/context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDebugContext(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("debug-breakpoint", "request example:*")
	req.Header.Add("debug-session", "http://example/123")
	session, _ := BuildSession("example", req.Header)

	ctx := AttachSession(context.Background(), session)
	session, ok := ExtractSession(ctx)
	if !ok {
		t.Errorf("session not found on context!")
	}
	if session.Name != "example" {
		t.Errorf("wrong name!")
	}
}

func TestResponseHook(t *testing.T) {
	// target of client request
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gores.String(w, 200, "hello world")
	}))
	defer ts.Close()

	// debug server transforming body
	ds := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		fmt.Fprint(w, `"body":"TRANSFORMED"`)
	}))
	defer ds.Close()

	// sadly, easiest to make session this way!
	// TODO make instantiating a session less convoluted!
	req, _ := http.NewRequest("GET", "/", nil)

	// TODO client breakpoint definitions are on call to, or call from?
	req.Header.Add("debug-breakpoint", "response example:*")
	req.Header.Add("debug-session", ds.URL)
	session, _ := BuildSession("example", req.Header)

	c := NewClient(http.DefaultClient)

	ctx := AttachSession(context.Background(), session)
	r, err := c.Get(ctx, ts.URL)
	if err != nil {
		t.Errorf("error issuing request: %s", err)
	}
	defer r.Body.Close()

	body, _ := ioutil.ReadAll(r.Body)
	if string(body) != "TRANSFORMED" {
		t.Errorf("expected body to be TRANSFORMED, it was '%s'", body)
	}

}
