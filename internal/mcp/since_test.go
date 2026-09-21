package mcp

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSinceCountsTheElapsedTime(t *testing.T) {
	h := testHandlers(time.Date(2026, time.September, 21, 22, 49, 3, 0, testLocation))

	result, out, err := h.since(context.Background(), nil, sinceInput{Time: "2026-09-21T19:49:03+10:00"})
	if err != nil {
		t.Fatalf("since() error: %v", err)
	}

	if out.Human != "3 hours ago" {
		t.Errorf("human = %q, want 3 hours ago", out.Human)
	}

	if out.Seconds != 10800 {
		t.Errorf("seconds = %d, want 10800", out.Seconds)
	}

	if out.At != "2026-09-21T19:49:03+10:00" {
		t.Errorf("at = %q, want 2026-09-21T19:49:03+10:00", out.At)
	}

	if out.Now != "2026-09-21T22:49:03+10:00" {
		t.Errorf("now = %q, want 2026-09-21T22:49:03+10:00", out.Now)
	}

	if text := singleText(t, result); !strings.Contains(text, out.Human) {
		t.Errorf("text = %q, want it to carry %q", text, out.Human)
	}
}

func TestSinceFutureTimestamp(t *testing.T) {
	h := testHandlers(time.Date(2026, time.September, 21, 22, 49, 3, 0, testLocation))

	_, out, err := h.since(context.Background(), nil, sinceInput{Time: "2026-09-22T01:49:03+10:00"})
	if err != nil {
		t.Fatalf("since() error: %v", err)
	}

	if out.Human != "in 3 hours" {
		t.Errorf("human = %q, want in 3 hours", out.Human)
	}

	if out.Seconds != -10800 {
		t.Errorf("seconds = %d, want -10800", out.Seconds)
	}
}

func TestSinceZoneLessTimestampUsesHandlerZone(t *testing.T) {
	h := testHandlers(time.Date(2026, time.September, 21, 22, 49, 3, 0, testLocation))

	_, out, err := h.since(context.Background(), nil, sinceInput{Time: "2026-09-21 19:49:03"})
	if err != nil {
		t.Fatalf("since() error: %v", err)
	}

	if out.At != "2026-09-21T19:49:03+10:00" {
		t.Errorf("at = %q, want the zone-less input read in the handler zone", out.At)
	}
}

func TestSinceZonedTimestampIgnoresHandlerZone(t *testing.T) {
	h := testHandlers(time.Date(2026, time.September, 21, 22, 49, 3, 0, testLocation))

	_, out, err := h.since(context.Background(), nil, sinceInput{Time: "2026-09-21T09:49:03Z"})
	if err != nil {
		t.Fatalf("since() error: %v", err)
	}

	if out.At != "2026-09-21T19:49:03+10:00" {
		t.Errorf("at = %q, want the same instant reported in the handler zone", out.At)
	}

	if out.Human != "3 hours ago" {
		t.Errorf("human = %q, want 3 hours ago", out.Human)
	}
}

func TestSinceRejectsNonTimestamp(t *testing.T) {
	h := testHandlers(time.Date(2026, time.September, 21, 22, 49, 3, 0, testLocation))

	_, _, err := h.since(context.Background(), nil, sinceInput{Time: "yesterday"})
	if err == nil || !strings.HasPrefix(err.Error(), "since:") {
		t.Fatalf("error = %v, want a since: prefix", err)
	}
}

func TestSinceCallTool(t *testing.T) {
	srv, err := New(Deps{Log: zapNop()})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	at := time.Now().Add(-2 * time.Hour).Format(time.RFC3339)

	result, err := connectSession(t, srv).CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "since",
		Arguments: map[string]any{"time": at},
	})
	if err != nil {
		t.Fatalf("CallTool() error: %v", err)
	}

	if result.IsError {
		t.Fatalf("CallTool() tool error: %+v", result.Content)
	}

	if !strings.Contains(singleText(t, result), "2 hours ago") {
		t.Errorf("content = %+v, want a two hour answer", result.Content)
	}
}
