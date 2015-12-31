package rpcdb
import (
"testing"
"net/http"
	"io/ioutil"
	"strings"
)

func TestRoundTripper(t *testing.T) {
	c := http.Client {
		Transport: FakeTransport{},
	}

	resp, _ := c.Get("http://www.example.com/")

	out, _ := ioutil.ReadAll(resp.Body)
	if string(out) != "fake hello!" {
		t.Errorf("expected 'fake hello!' got %s", out)
	}

}

type FakeTransport struct {

}

func (f FakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp := http.Response{
		Body: ioutil.NopCloser(strings.NewReader("fake hello!")),
	}
	return &resp, nil
}

func (f FakeTransport) CancelRequest(req *http.Request) {
	// noop
}
