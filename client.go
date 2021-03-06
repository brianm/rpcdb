package rpcdb

import (
	"fmt"
	"golang.org/x/net/context"
	"io"
	"net/http"
)

// TODO use proper ctxhttp package
// 		need this to fiddle timeouts properly
// TODO refactor client to use arbitrary interceptors, with debug as an interceptor
// TODO move ^^ generic client into its own package

const sessionKey = "github.com/brianm/rpcdb:debug_session_key"

func AttachSession(ctx context.Context, session Session) context.Context {
	return context.WithValue(ctx, sessionKey, session)
}

func ExtractSession(ctx context.Context) (Session, bool) {
	s, ok := ctx.Value(sessionKey).(Session)
	return s, ok
}

type DebugClient struct {
	http *http.Client
}

func NewClient(hc *http.Client) DebugClient {
	return DebugClient{hc}
}

func (c DebugClient) Do(ctx context.Context, req *http.Request) (*http.Response, error) {
	session, ok := ExtractSession(ctx)

	newReq, err := session.Request(req)
	if err != nil {
		return nil, err
	}

	resp, err := c.http.Do(newReq)
	if err != nil {
		return resp, err
	}
	if ok {
		return session.Response(req, resp)
	} else {
		return resp, nil
	}
}

func (c DebugClient) Get(ctx context.Context, url string) (resp *http.Response, err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("unable to create request: %s", err)
	}
	return c.Do(ctx, req)
}

func (c DebugClient) Post(ctx context.Context, url string, bodyType string, body io.Reader) (resp *http.Response, err error) {
	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, fmt.Errorf("unable to make request: %s", err)
	}
	req.Header.Add("Content-Type", bodyType)
	return c.Do(ctx, req)
}
