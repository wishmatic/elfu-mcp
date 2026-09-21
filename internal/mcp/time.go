package mcp

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wishmatic/elfu-mcp/internal/timing"
	"go.uber.org/zap"
)

type timeInput struct{}

type timeOutput struct {
	ISO      string `json:"iso" jsonschema:"the current time as an ISO 8601 timestamp with its UTC offset"`
	Unix     int64  `json:"unix" jsonschema:"the current time as seconds since the Unix epoch"`
	Zone     string `json:"zone" jsonschema:"the timezone the timestamp is reported in"`
	Offset   string `json:"offset" jsonschema:"the UTC offset of that timezone, e.g. +10:00"`
	Readable string `json:"readable" jsonschema:"the current time spelled out for people"`
}

func registerTime(srv *mcp.Server, h *handlers) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "time",
		Description: "Report the current date and time, in the server's timezone and with its UTC offset. Call it " +
			"whenever the current time matters, rather than assuming today's date or the time of day.",
		InputSchema: inputSchema[timeInput]("time"),
	}, h.currentTime)
}

func (h *handlers) currentTime(
	_ context.Context,
	_ *mcp.CallToolRequest,
	_ timeInput,
) (*mcp.CallToolResult, timeOutput, error) {
	stamp := timing.Stamp{At: h.now().In(h.location)}

	out := timeOutput{
		ISO:      stamp.ISO(),
		Unix:     stamp.Unix(),
		Zone:     stamp.Zone(),
		Offset:   stamp.Offset(),
		Readable: stamp.Readable(),
	}

	h.log.Info("time reported", zap.String("iso", out.ISO), zap.String("zone", out.Zone))

	text := fmt.Sprintf("Current time: %s (%s, Unix %d).", out.ISO, out.Readable, out.Unix)

	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}, out, nil
}
