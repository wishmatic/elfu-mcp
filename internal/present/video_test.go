package present

import (
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const testVideoURL = "https://cdn.example.com/clip.mp4"

func TestEmbedVideoContent(t *testing.T) {
	content := EmbedVideo(testVideoURL)
	if len(content) != 2 {
		t.Fatalf("content = %d, want a caption and one resource", len(content))
	}

	caption, ok := content[0].(*mcp.TextContent)
	if !ok || !strings.Contains(caption.Text, testVideoURL) {
		t.Fatalf("content[0] = %#v, want a caption naming the URL", content[0])
	}

	resource := embeddedVideo(t, content)

	if !strings.HasPrefix(resource.URI, "ui://") {
		t.Errorf("uri = %q, want a ui:// URI", resource.URI)
	}

	if resource.MIMEType != "text/html" {
		t.Errorf("mime type = %q, want text/html", resource.MIMEType)
	}
}

func TestEmbedVideoPlayer(t *testing.T) {
	html := embeddedVideo(t, EmbedVideo(testVideoURL)).Text

	if !strings.Contains(html, `<video src="`+testVideoURL+`"`) {
		t.Errorf("resource text = %q, want a video element for the URL", html)
	}

	for _, attribute := range []string{"controls", "playsinline", `preload="metadata"`} {
		if !strings.Contains(html, attribute) {
			t.Errorf("resource text = %q, want %s", html, attribute)
		}
	}

	for _, attribute := range []string{"autoplay", "muted"} {
		if strings.Contains(html, attribute) {
			t.Errorf("resource text = %q, want no %s", html, attribute)
		}
	}
}

func TestEmbedVideoEscapesURL(t *testing.T) {
	hostile := `https://cdn.example.com/x.mp4?a=1&b="><script>alert(1)</script>`
	html := embeddedVideo(t, EmbedVideo(hostile)).Text

	if strings.Contains(html, "<script>") {
		t.Fatalf("resource text = %q, want the URL escaped, not markup", html)
	}

	if !strings.Contains(html, "&lt;script&gt;") || !strings.Contains(html, "&amp;b=") {
		t.Errorf("resource text = %q, want the URL escaped", html)
	}
}

func TestEmbedVideoURIFollowsURL(t *testing.T) {
	uri := embeddedVideo(t, EmbedVideo(testVideoURL)).URI

	if uri != embeddedVideo(t, EmbedVideo(testVideoURL)).URI {
		t.Errorf("uri = %q, want the same URI for a repeated URL", uri)
	}

	if uri == embeddedVideo(t, EmbedVideo("https://cdn.example.com/other.mp4")).URI {
		t.Errorf("uri = %q, want a distinct URI per URL", uri)
	}
}

func embeddedVideo(t *testing.T, content []mcp.Content) *mcp.ResourceContents {
	t.Helper()

	if len(content) != 2 {
		t.Fatalf("content = %d, want a caption and one resource", len(content))
	}

	resource, ok := content[1].(*mcp.EmbeddedResource)
	if !ok {
		t.Fatalf("content[1] = %#v, want an embedded resource", content[1])
	}

	return resource.Resource
}
