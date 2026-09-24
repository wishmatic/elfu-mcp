package mcp

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wishmatic/elfu-mcp/internal/timing"
	"go.uber.org/zap"
)

// maxSleep caps one call, so a model cannot hold a request open for as long as it likes.
const maxSleep = 30 * time.Second

type sleepInput struct {
	Seconds float64 `json:"seconds" jsonschema:"how long to wait, in seconds; 30 is the most a single call will wait"`
}

type sleepOutput struct {
	Seconds float64 `json:"seconds" jsonschema:"how long the call waited, in seconds"`
	Now     string  `json:"now" jsonschema:"the time the call returned, as an ISO 8601 timestamp with its UTC offset"`
}

func registerSleep(srv *mcp.Server, h *handlers) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "sleep",
		Description: "Wait for a number of seconds, up to 30, and then report the time it woke. Use it when something " +
			"else needs a moment to finish, such as a file that is still being written, rather than retrying at once.",
		InputSchema: inputSchema[sleepInput]("sleep"),
		Annotations: toolAnnotations(false),
	}, h.sleep)
}

func (h *handlers) sleep(
	ctx context.Context,
	_ *mcp.CallToolRequest,
	in sleepInput,
) (*mcp.CallToolResult, sleepOutput, error) {
	wait, err := sleepDuration(in.Seconds)
	if err != nil {
		return nil, sleepOutput{}, err
	}

	if err := waitFor(ctx, wait); err != nil {
		h.log.Warn("sleep interrupted", zap.Float64("seconds", in.Seconds), zap.Error(err))

		return nil, sleepOutput{}, fmt.Errorf("sleep: %w", err)
	}

	out := sleepOutput{
		Seconds: in.Seconds,
		Now:     timing.Stamp{At: h.now().In(h.location)}.ISO(),
	}

	h.log.Info("slept", zap.Float64("seconds", out.Seconds))

	text := fmt.Sprintf("Slept %gs; the server now reads %s.", out.Seconds, out.Now)

	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}, out, nil
}

func sleepDuration(seconds float64) (time.Duration, error) {
	if math.IsNaN(seconds) || seconds < 0 {
		return 0, fmt.Errorf("sleep: seconds must be zero or more, got %v", seconds)
	}

	if seconds > maxSleep.Seconds() {
		return 0, fmt.Errorf("sleep: seconds must be at most %g, got %v", maxSleep.Seconds(), seconds)
	}

	return time.Duration(seconds * float64(time.Second)), nil
}

func waitFor(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-timer.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
