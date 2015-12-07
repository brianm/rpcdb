package rpcdb

import (
	"testing"
)

func TestParseReceiveExpr(t *testing.T) {
	bp, err := ParseExpression("receive example:/hello")
	if err != nil {
		t.Errorf("failed tp parse 'receive example:/hello': %s", err)
	}

	if bp.Hook != Receive {
		t.Errorf("expected receive hook, got %s", bp.Hook)
	}
	if bp.ServiceName != "example" {
		t.Errorf("expected service=example, got service=%s", bp.Hook)
	}
	if bp.RPCName != "/hello" {
		t.Errorf("expected rpc=/hello, got rpc=%s", bp.Hook)
	}
}

func TestParsePattern(t *testing.T) {
	parts := parsePattern.FindStringSubmatch("receive example:*")
	if len(parts) != 4 {
		t.Errorf("unexpected matched parts! %v", parts)
	}

	if parts[1] != "receive" {
		t.Errorf("didn't parse out receive properly")
	}

	if parts[2] != "example" {
		t.Errorf("didn't parse out example properly")
	}

	if parts[3] != "*" {
		t.Errorf("didn't parse out * correctly")
	}
}
