package rpcdb

import (
	"bytes"
	"fmt"
	"go/src/io/ioutil"
	"go/src/strings"
	"net/http"
	"regexp"

	"github.com/docker/docker/vendor/src/github.com/jfrazelle/go/canonical/json"
)

var debugBreakpointHeaderKey = http.CanonicalHeaderKey("Debug-Breakpoint")
var debugSessionHeaderKey = http.CanonicalHeaderKey("Debug-Session")

// HookType is the type of hook
type HookType int

func (h HookType) String() string {
	switch h {
	case Receive:
		return "receive"
	case Reply:
		return "reply"
	case Request:
		return "request"
	case Response:
		return "response"
	}
	panic("unexpected HookType")
}

// ParseHookType parses a hook name to the appropriate HookType
func ParseHookType(name string) (HookType, error) {
	switch name {
	case "receive":
		return Receive, nil
	case "reply":
		return Reply, nil
	case "request":
		return Request, nil
	case "response":
		return Response, nil
	}
	return Receive, fmt.Errorf("unknown hook type: %s", name)
}

const (
	// Receive HookType
	Receive HookType = iota
	// Reply HookType
	Reply
	// Request HookType
	Request
	// Response HookType
	Response
)

var parsePattern = regexp.MustCompile(`(\w+)\s+(\w+)\:(.+)`)

// ParseExpression parses a single breakpoint expression
func ParseExpression(expr string) (Breakpoint, error) {
	bp := Breakpoint{}
	parts := parsePattern.FindStringSubmatch(expr)
	if len(parts) != 4 {
		return bp, fmt.Errorf("unable to parse breakpoint expression '%s'", expr)
	}

	hookType, err := ParseHookType(parts[1])
	if err != nil {
		return bp, err
	}
	bp.Hook = hookType
	bp.ServiceName = parts[2]
	bp.RPCName = parts[3]
	return bp, nil
}

// Breakpoint represents the parsed breakpoint expression
type Breakpoint struct {
	Hook        HookType
	ServiceName string
	RPCName     string
}

// Breakpoints is a broken out view of found breakpoints
type Session struct {
	Name                string
	SessionURL          string
	ReceiveBreakpoints  []Breakpoint
	ReplyBreakpoints    []Breakpoint
	RequestBreakpoints  []Breakpoint
	ResponseBreakpoints []Breakpoint
}

// Receive should be called to exercise any receive break points
func (s Session) Receive(req *http.Request) (*http.Request, error) {
	for _, bp := range s.ReceiveBreakpoints {
		// TODO handle wildcard service name matches
		if bp.ServiceName == s.Name {
			// TODO handle wildcard endpoint matches
			if bp.RPCName == req.URL.Path {
				requestBody, err := ioutil.ReadAll(req.Body)
				if err != nil {
					return nil, fmt.Errorf("error reading body: %s", err)
				}

				resp, err := http.Post(s.SessionURL, "text/plain", bytes.NewReader(requestBody))
				if err != nil {
					return nil, fmt.Errorf("error calling debugger: %s", err)
				}
				defer resp.Body.Close()

				debugResponseBody, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					return nil, fmt.Errorf("Unable to read debugger response: %s", err)
				}
				resp.Body.Close()

				rb := ReceiveBody{}
				err = json.Unmarshal(debugResponseBody, &rb)
				if err != nil {
					return nil, fmt.Errorf("Unable to parse debugger response: %s", err)
				}

				newReq, err := http.NewRequest(req.Method, req.URL.String(), strings.NewReader(rb.Body))
				if err != nil {
					return nil, fmt.Errorf("unable to construct replacement body from debugger: %s", err)
				}

				return newReq, nil
			}
		}
	}

	return req, nil
}

// BuildSession builds a session from http header information
func BuildSession(name string, req http.Header) (Session, error) {
	session := Session{
		Name:       name,
		SessionURL: req.Get(debugSessionHeaderKey),
	}
	breakpoints := req[debugBreakpointHeaderKey]
	for _, expr := range breakpoints {
		bp, err := ParseExpression(expr)
		if err != nil {
			return session, err
		}
		switch bp.Hook {
		case Receive:
			session.ReceiveBreakpoints = append(session.ReceiveBreakpoints, bp)
		case Reply:
			session.ReplyBreakpoints = append(session.ReplyBreakpoints, bp)
		case Request:
			session.RequestBreakpoints = append(session.RequestBreakpoints, bp)
		case Response:
			session.ResponseBreakpoints = append(session.ResponseBreakpoints, bp)
		}
	}

	return session, nil
}

type ReceiveBody struct {
	Body string
}
