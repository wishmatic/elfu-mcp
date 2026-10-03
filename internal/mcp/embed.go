package mcp

import (
	"context"
	"errors"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wishmatic/elfu-mcp/internal/present"
	"go.uber.org/zap"
)

type embedInput struct {
	VideoURL string `json:"video_url" jsonschema:"URL of the video to embed; the user's browser loads it, so it must be reachable without the chat's credentials"`
}

type embedOutput struct {
	URL string `json:"url" jsonschema:"the video URL embedded in the chat"`
}

func registerEmbed(srv *mcp.Server, h *handlers) {
	mcp.AddTool(srv, &mcp.Tool{
		Name: "embed",
		Description: "Embed a video into the chat so the user can play it inline. Nothing is downloaded: the user's " +
			"browser loads the URL itself, so pass a direct link to a video file the browser can reach on its own. A " +
			"link to a page that merely contains a video will not play.",
		InputSchema: inputSchema[embedInput]("embed"),
		Annotations: toolAnnotations(true),
	}, h.embed)
}

func (h *handlers) embed(
	_ context.Context,
	_ *mcp.CallToolRequest,
	in embedInput,
) (*mcp.CallToolResult, embedOutput, error) {
	videoURL := strings.TrimSpace(in.VideoURL)
	if videoURL == "" {
		return nil, embedOutput{}, errors.New("embed: a video URL is required")
	}

	h.log.Info("embedded video", zap.String("video_url", videoURL))

	return &mcp.CallToolResult{Content: present.EmbedVideo(videoURL)}, embedOutput{URL: videoURL}, nil
}
