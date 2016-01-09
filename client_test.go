package rpcdb
import (
	"testing"
	"net/http"
	"golang.org/x/net/context"
)

func TestDebugContext(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	req.Header.Add("debug-breakpoint", "request example:*")
	req.Header.Add("debug-session", "http://example/123")

	ctx, _ := AttachSession(context.Background(), "example", req)
	session, ok := ExtractSession(ctx)
	if !ok {
		t.Errorf("session not found on context!")
	}
	if session.Name != "example" {
		t.Errorf("wrong name!")
	}
}
