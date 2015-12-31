package rpcdb
import (
	"testing"
	"net/http"
	"io/ioutil"
	"strings"
	"fmt"
	"net/http/httptest"
	"bytes"
)

func TestFakeRoundTripper(t *testing.T) {
	c := http.Client{
		Transport: FakeTransport{},
	}

	resp, _ := c.Get("http://www.example.com/")

	out, _ := ioutil.ReadAll(resp.Body)
	if string(out) != "fake hello!" {
		t.Errorf("expected 'fake hello!' got %s", out)
	}
}

type FakeTransport struct{}

func (f FakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := http.Response{
		Body: ioutil.NopCloser(strings.NewReader("fake hello!")),
	}
	return &resp, nil
}


func TestDebugRequest(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/plain")
		fmt.Fprintln(w, "TRANSFORMED")
	}))
	defer target.Close()

	debugger := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "text/plain")
		fmt.Fprintln(w, "TRANSFORMED")
	}))
	defer debugger.Close()

	client := http.Client{
		Transport: NewTransport("example", nil),
	}

	req, _ := http.NewRequest("POST", target.URL, bytes.NewReader([]byte("hello world")))

	resp, err := client.Do(req)
	if err != nil {
		t.Errorf("error executing request: %s", err)
	}

}
