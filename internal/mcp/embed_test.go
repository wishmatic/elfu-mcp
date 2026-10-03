package mcp

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const testVideoURL = "https://cdn.example.com/clip.mp4"

func TestEmbedCallTool(t *testing.T) {
	srv, err := New(Deps{Log: zapNop()})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	result := callEmbed(t, srv, testVideoURL)

	if len(result.Content) != 2 {
		t.Fatalf("content = %d, want a caption and one resource", len(result.Content))
	}

	caption, ok := result.Content[0].(*mcp.TextContent)
	if !ok || !strings.Contains(caption.Text, testVideoURL) {
		t.Fatalf("content[0] = %#v, want a caption naming the URL", result.Content[0])
	}

	resource, ok := result.Content[1].(*mcp.EmbeddedResource)
	if !ok {
		t.Fatalf("content[1] = %#v, want an embedded resource", result.Content[1])
	}

	if !strings.HasPrefix(resource.Resource.URI, "ui://") {
		t.Errorf("uri = %q, want a ui:// URI", resource.Resource.URI)
	}

	if resource.Resource.MIMEType != "text/html" {
		t.Errorf("mime type = %q, want text/html", resource.Resource.MIMEType)
	}

	if !strings.Contains(resource.Resource.Text, `<video src="`+testVideoURL+`"`) {
		t.Errorf("resource text = %q, want a video element for the URL", resource.Resource.Text)
	}
}

func TestEmbedDoesNotFetchVideo(t *testing.T) {
	var requests atomic.Int64

	video := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		http.Error(w, "embed must not fetch", http.StatusInternalServerError)
	}))
	t.Cleanup(video.Close)

	srv, err := New(Deps{Log: zapNop(), Resolver: newResolver(t)})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	callEmbed(t, srv, video.URL+"/clip.mp4")

	if got := requests.Load(); got != 0 {
		t.Errorf("requests = %d, want none", got)
	}
}

func TestEmbedRejectsEmptyURL(t *testing.T) {
	srv, err := New(Deps{Log: zapNop()})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	result, err := connectSession(t, srv).CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "embed",
		Arguments: map[string]any{"video_url": "   "},
	})
	if err != nil {
		t.Fatalf("CallTool() error: %v", err)
	}

	if !result.IsError {
		t.Fatalf("CallTool() content = %+v, want a tool error", result.Content)
	}
}

func TestEmbedRequiresURL(t *testing.T) {
	srv, err := New(Deps{Log: zapNop()})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	result, err := connectSession(t, srv).CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "embed",
		Arguments: map[string]any{},
	})
	if err != nil {
		t.Fatalf("CallTool() error: %v", err)
	}

	if !result.IsError {
		t.Fatalf("CallTool() content = %+v, want a tool error", result.Content)
	}
}

func callEmbed(t *testing.T, srv *mcp.Server, videoURL string) *mcp.CallToolResult {
	t.Helper()

	result, err := connectSession(t, srv).CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "embed",
		Arguments: map[string]any{"video_url": videoURL},
	})
	if err != nil {
		t.Fatalf("CallTool() error: %v", err)
	}

	if result.IsError {
		t.Fatalf("CallTool() tool error: %+v", result.Content)
	}

	return result
}
