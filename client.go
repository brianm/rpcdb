package rpcdb
import (
	"net/http"
	"fmt"
	"golang.org/x/net/context"
	"io"
)

const sessionKey = "github.com/brianm/rpcdb:debug_session_key"

type Client struct {
	http *http.Client
}

func Wrap(hc *http.Client) Client {
	return Client{hc}
}

func (c Client) Do(ctx context.Context, req *http.Request) (resp *http.Response, err error) {
	return c.http.Do(req)
}

func (c Client) Get(ctx context.Context, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %s", err)
	}
	return c.Do(ctx, req)
}

func (c Client) Post(ctx context.Context, url string, bodyType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("unable to make request: %s", err)
	}
	req.Header.Add("Content-Type", bodyType)
	return c.Do(ctx, req)
}

func AttachSession(ctx context.Context, name string, req *http.Request) (context.Context, error) {
	session, err := BuildSession(name, req.Header)
	if err != nil {
		return ctx, err
	}
	return context.WithValue(ctx, sessionKey, session), nil
}

func ExtractSession(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(sessionKey).(Session)
	return s, ok
}

