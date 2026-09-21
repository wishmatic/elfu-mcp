package mcp

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestTimeReportsThePinnedInstant(t *testing.T) {
	at := time.Date(2026, time.September, 21, 22, 49, 3, 0, testLocation)

	result, out, err := testHandlers(at).currentTime(context.Background(), nil, timeInput{})
	if err != nil {
		t.Fatalf("currentTime() error: %v", err)
	}

	if out.ISO != "2026-09-21T22:49:03+10:00" {
		t.Errorf("iso = %q, want 2026-09-21T22:49:03+10:00", out.ISO)
	}

	if out.Unix != at.Unix() {
		t.Errorf("unix = %d, want %d", out.Unix, at.Unix())
	}

	if out.Zone != "AEST" {
		t.Errorf("zone = %q, want AEST", out.Zone)
	}

	if out.Offset != "+10:00" {
		t.Errorf("offset = %q, want +10:00", out.Offset)
	}

	if out.Readable != "Monday, 21 September 2026 at 22:49 AEST" {
		t.Errorf("readable = %q, want Monday, 21 September 2026 at 22:49 AEST", out.Readable)
	}

	text := singleText(t, result)
	if !strings.Contains(text, out.ISO) || !strings.Contains(text, out.Zone) {
		t.Errorf("text = %q, want it to carry the timestamp and the zone", text)
	}
}

func TestTimeConvertsToTheHandlerZone(t *testing.T) {
	utc := time.Date(2026, time.September, 21, 12, 49, 3, 0, time.UTC)

	_, out, err := testHandlers(utc).currentTime(context.Background(), nil, timeInput{})
	if err != nil {
		t.Fatalf("currentTime() error: %v", err)
	}

	if out.ISO != "2026-09-21T22:49:03+10:00" {
		t.Errorf("iso = %q, want the instant reported in the handler zone", out.ISO)
	}
}

func TestTimeCallTool(t *testing.T) {
	srv, err := New(Deps{Log: zapNop()})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	result, err := connectSession(t, srv).CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "time",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool() error: %v", err)
	}

	if result.IsError {
		t.Fatalf("CallTool() tool error: %+v", result.Content)
	}

	if !strings.Contains(singleText(t, result), "Current time") {
		t.Errorf("content = %+v, want a current time note", result.Content)
	}
}

func singleText(t *testing.T, result *mcp.CallToolResult) string {
	t.Helper()

	if len(result.Content) != 1 {
		t.Fatalf("content = %d, want a single text block", len(result.Content))
	}

	text, ok := result.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("content[0] = %#v, want text", result.Content[0])
	}

	return text.Text
}
