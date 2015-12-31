package rpcdb

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"encoding/json"
	"net/http/httptest"
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

// BuildSession builds a session from http header information
func BuildSession(name string, header http.Header) (Session, error) {
	session := Session{
		Name:       name,
		SessionURL: header.Get(debugSessionHeaderKey),
	}
	breakpoints := header[debugBreakpointHeaderKey]
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

// Receive should be called to exercise any receive break points
func (s Session) Receive(req *http.Request) (*http.Request, error) {
	// TODO this nested for/if/if is repeated for every BP match test, refactor to common function
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

				for k, vs := range req.Header {
					for _, v := range vs {
						newReq.Header.Add(k, v)
					}
				}

				// return from inside loop so that we only trigger once, even if
				// multiple breakpoint definitions match
				return newReq, nil
			}
		}
	}

	return req, nil
}

type ReceiveBody struct {
	Body string
}

func (s Session) StartReply(w http.ResponseWriter, req *http.Request) ReplyTrap {
	// default config is to not capture, if we find a relevant BP we convert to capture
	rep := ReplyTrap{
		writer:    w,
		debugging: false,
		recorder:  httptest.NewRecorder(),
		session:   s,
	}

	// TODO this nested for/if/if is repeated for every BP match test, refactor to common function
	for _, bp := range s.ReplyBreakpoints {
		// TODO handle wildcard service name matches
		if bp.ServiceName == s.Name {
			// TODO handle wildcard endpoint matches
			if bp.RPCName == req.URL.Path {
				rep.debugging = true
				return rep
			}
		}
	}
	return rep
}

// ReplyTrap captures a server reply in order to send it to the
// debugger for consideration. It operates in two parts, first is
// capture, second is acting on what was captured. To do the "act on"
// part `FinishReply` must be invoked.
type ReplyTrap struct {
	writer    http.ResponseWriter
	recorder  *httptest.ResponseRecorder
	debugging bool
	session   Session
}

// CaptureWriter returns the resposne writer to be used to capture the
// server reply
func (r ReplyTrap) CaptureWriter() http.ResponseWriter {
	if r.debugging {
		return r.recorder
	} else {
		return r.writer
	}
}

// FinishReply sends the captured reply to the debugger, if needed, and
// sends anything needed out to on the real reply. If there is no breakpoint
// on the reply this is a no-op
func (r ReplyTrap) FinishReply() error {
	if r.debugging {
		// r.recorder has the actual recorded response, now we need to
		// send it to the debugger

		// just send the body for now, will flesh out debugger protocol
		// once we validate capture of all four BP types is viable
		resp, err := http.Post(r.session.SessionURL, "text/plain", r.recorder.Body)
		if err != nil {
			return fmt.Errorf("Error making request to debugger: %s: err")
		}

		debugReplyBody, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("unable to read response from debugger: %s", err)
		}

		debuggerResponse := ReplyBody{}
		err = json.Unmarshal(debugReplyBody, &debuggerResponse)
		if err != nil {
			return fmt.Errorf("unable to parse response from debugger: %s", err)
		}

		// copy response directly from the recorder to the real response
		// just as filler for now :-)
		hdr := r.writer.Header()
		for k, vs := range hdr {
			for _, v := range vs {
				r.writer.Header().Add(k, v)
			}
		}
		r.writer.WriteHeader(r.recorder.Code)

		/*
			// code to copy
			_, err = io.Copy(r.writer, r.recorder.Body)
			if err != nil {
				return fmt.Errorf("Error copying recorded body: %s", err)
			}
		*/
		r.writer.Write([]byte(debuggerResponse.Body))
	}
	return nil
}

type ReplyBody struct {
	Body string
}


func (s Session) Request(req *http.Request) (*http.Request, error) {
	return req, nil
}

func (s Session) Response(resp *http.Response) (*http.Response, error) {
	return resp, nil
}
