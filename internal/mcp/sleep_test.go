package mcp

import (
	"context"
	"math"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestSleepWaitsAndReportsTheWakeTime(t *testing.T) {
	at := time.Date(2026, time.September, 21, 22, 49, 3, 0, testLocation)

	start := time.Now()
	result, out, err := testHandlers(at).sleep(context.Background(), nil, sleepInput{Seconds: 0.02})
	if err != nil {
		t.Fatalf("sleep() error: %v", err)
	}

	if waited := time.Since(start); waited < 20*time.Millisecond {
		t.Errorf("call returned after %s, want it to have waited 20ms", waited)
	}

	if out.Seconds != 0.02 {
		t.Errorf("seconds = %v, want 0.02", out.Seconds)
	}

	if out.Now != "2026-09-21T22:49:03+10:00" {
		t.Errorf("now = %q, want the handler's own clock", out.Now)
	}

	if text := singleText(t, result); !strings.Contains(text, "Slept") || !strings.Contains(text, out.Now) {
		t.Errorf("text = %q, want it to name the wait and the time it woke", text)
	}
}

func TestSleepRejectsWaitsBeyondTheCap(t *testing.T) {
	tests := []struct {
		name    string
		seconds float64
	}{
		{name: "beyond the cap", seconds: 31},
		{name: "an hour", seconds: 3600},
		{name: "an infinity", seconds: math.Inf(1)},
		{name: "not a number", seconds: math.NaN()},
		{name: "negative", seconds: -1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			start := time.Now()

			_, _, err := testHandlers(time.Now()).sleep(context.Background(), nil, sleepInput{Seconds: tt.seconds})
			if err == nil || !strings.HasPrefix(err.Error(), "sleep:") {
				t.Fatalf("error = %v, want a sleep: prefix", err)
			}

			if waited := time.Since(start); waited > time.Second {
				t.Errorf("refusal took %s, want it to refuse without waiting", waited)
			}
		})
	}
}

func TestSleepAllowsTheCap(t *testing.T) {
	if _, err := sleepDuration(maxSleep.Seconds()); err != nil {
		t.Errorf("sleepDuration(%g) error: %v", maxSleep.Seconds(), err)
	}

	if _, err := sleepDuration(0); err != nil {
		t.Errorf("sleepDuration(0) error: %v", err)
	}
}

func TestSleepStopsWhenTheCallIsCancelled(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	start := time.Now()

	_, _, err := testHandlers(time.Now()).sleep(ctx, nil, sleepInput{Seconds: 30})
	if err == nil || !strings.HasPrefix(err.Error(), "sleep:") {
		t.Fatalf("error = %v, want a sleep: prefix", err)
	}

	if waited := time.Since(start); waited > 5*time.Second {
		t.Errorf("call returned after %s, want it to stop with the context", waited)
	}
}

func TestSleepCallTool(t *testing.T) {
	srv, err := New(Deps{Log: zapNop()})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	result, err := connectSession(t, srv).CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "sleep",
		Arguments: map[string]any{"seconds": 0.01},
	})
	if err != nil {
		t.Fatalf("CallTool() error: %v", err)
	}

	if result.IsError {
		t.Fatalf("CallTool() tool error: %+v", result.Content)
	}

	if !strings.Contains(singleText(t, result), "Slept") {
		t.Errorf("content = %+v, want a note about the wait", result.Content)
	}
}

func TestSleepCallToolRejectsBeyondTheCap(t *testing.T) {
	srv, err := New(Deps{Log: zapNop()})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	result, err := connectSession(t, srv).CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "sleep",
		Arguments: map[string]any{"seconds": 600},
	})
	if err != nil {
		t.Fatalf("CallTool() error: %v", err)
	}

	if !result.IsError {
		t.Fatalf("CallTool() = %+v, want a tool error naming the cap", result.Content)
	}

	if text := singleText(t, result); !strings.Contains(text, "30") {
		t.Errorf("text = %q, want it to name the cap", text)
	}
}
