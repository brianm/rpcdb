package rpcdb

import (
	"fmt"
	"net/http"
	"regexp"
)

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

// ParseBreakpoints breakpoint expressions out of HTTP Headers
func ParseBreakpoints(req http.Header) ([]Breakpoint, error) {
	// type Header map[string][]string
	var resp []Breakpoint

	breakpoints := req[http.CanonicalHeaderKey("Debug-Breakpoint")]
	for _, expr := range breakpoints {
		bp, err := ParseExpression(expr)
		if err != nil {
			return resp, err
		}
		resp = append(resp, bp)
	}

	return resp, nil
}
