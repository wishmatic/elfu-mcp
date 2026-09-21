package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wishmatic/elfu-mcp/internal/timing"
	"go.uber.org/zap"
)

type sinceInput struct {
	Time string `json:"time" jsonschema:"a timestamp in ISO 8601, e.g. 2026-09-21T19:49:03+10:00 or 2026-09-21; one without a zone is read in the server's timezone"`
}

type sinceOutput struct {
	Human   string `json:"human" jsonschema:"how long ago that was, in words, e.g. 3 hours ago, or in 3 hours when the timestamp has not happened yet"`
	Seconds int64  `json:"seconds" jsonschema:"how long ago that was in seconds, negative when the timestamp is in the future"`
	At      string `json:"at" jsonschema:"the input timestamp, reported in the server's timezone"`
	Now     string `json:"now" jsonschema:"the current time the answer is relative to"`
}

func registerSince(srv *mcp.Server, h *handlers) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "since",
		Description: "Report how long ago an ISO 8601 timestamp was, in words and in seconds. Use it to make sense " +
			"of a timestamp in a message, a log, or a file, rather than estimating the elapsed time.",
		InputSchema: inputSchema[sinceInput]("since"),
	}, h.since)
}

func (h *handlers) since(
	_ context.Context,
	_ *mcp.CallToolRequest,
	in sinceInput,
) (*mcp.CallToolResult, sinceOutput, error) {
	then, err := timing.Parse(in.Time, h.location)
	if err != nil {
		h.log.Warn("since parse failed", zap.String("time", in.Time), zap.Error(err))

		return nil, sinceOutput{}, fmt.Errorf("since: parse time: %w", err)
	}

	now := h.now().In(h.location)
	then = then.In(h.location)

	out := sinceOutput{
		Human:   timing.HumanSince(then, now),
		Seconds: int64(now.Sub(then).Seconds()),
		At:      timing.Stamp{At: then}.ISO(),
		Now:     timing.Stamp{At: now}.ISO(),
	}

	h.log.Info("since reported", zap.String("at", out.At), zap.String("human", out.Human))

	text := fmt.Sprintf("%s (%s, with the server now at %s).", out.Human, out.At, out.Now)

	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}, out, nil
}
