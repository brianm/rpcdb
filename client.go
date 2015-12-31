package rpcdb
import (
	"net/http"
	"fmt"
)


// DebugTransport is an http.RoundTripper which supports rpcdb. It breaks the guidance
// on RoundTripper, which reads:
//
//     RoundTrip should not attempt to handle higher-level
//     protocol details such as redirects, authentication,
//     or cookies.
//
//     RoundTrip should not modify the request, except for
//     consuming and closing the Body, including on errors. The
//     request's URL and Header fields are guaranteed to be
//     initialized.
//
// However, I am unable to find a reasonable way of accomplishing this without breaking
// that guidance. Apologies in advance if this causes unexpected issues!
type DebugTransport struct {
	Name      string
	Transport http.RoundTripper
}

// NewTransport builds a new DebugTransport, wrapping an existing round tripper, if
// the transport is nil, uses http.DefaultTransport
func NewTransport(name string, transport http.RoundTripper) DebugTransport {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return DebugTransport{
		Name: name,
		Transport: transport,
	}
}

// RoundTrip provides the http.RoundTripper implementation which short circuits it as needed
// to provide debug functionality
func (t DebugTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// TODO fail quietly, log, or fail hard -- which is correct here, in all cases?
	//      for now, filing hard

	session, err := BuildSession(t.Name, req.Header)
	if err != nil {
		return nil, fmt.Errorf("Error building debug session: %s", err)
	}

	// handle request
	newReq, err := session.Request(req)
	if err != nil {
		return nil, fmt.Errorf("Error debugging request: %s", err)
	}


	// handle response
	resp, err := t.Transport.RoundTrip(newReq)
	newResp, err := session.Response(resp)
	if err != nil {
		return nil, fmt.Errorf("Error debugging response: %s", err)
	}

	return newResp, nil
}


func (t DebugTransport) CancelRequest(req *http.Request) {
	// implementation cribbed and adapted from net/http
	type canceler interface {
		CancelRequest(*http.Request)
	}
	tr, ok := t.Transport.(canceler)
	if ok {
		tr.CancelRequest(req)
	}
}
