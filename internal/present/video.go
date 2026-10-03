package present

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"html"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// EmbedVideo presents the video at videoURL as a caption and an HTML resource LibreChat renders inline.
//
// Nothing is fetched: the resource points the browser at videoURL, so it must be a URL the browser can load on its own.
func EmbedVideo(videoURL string) []mcp.Content {
	return []mcp.Content{
		&mcp.TextContent{Text: fmt.Sprintf("Embedded video from %s.", videoURL)},
		&mcp.EmbeddedResource{Resource: &mcp.ResourceContents{
			URI:      videoResourceURI(videoURL),
			MIMEType: "text/html",
			Text:     videoHTML(videoURL),
		}},
	}
}

// videoResourceURI is distinct per URL so a client that keys resources by URI cannot confuse two videos with each
// other; LibreChat keys them by the hash of the resource text and only checks the ui:// prefix here.
func videoResourceURI(videoURL string) string {
	sum := sha256.Sum256([]byte(videoURL))

	return "ui://elfu-mcp/video/" + hex.EncodeToString(sum[:6])
}

func videoHTML(videoURL string) string {
	return strings.Replace(videoHTMLTemplate, videoURLPlaceholder, html.EscapeString(videoURL), 1)
}

const videoURLPlaceholder = "{video_url}"

// videoHTMLTemplate is filled by videoHTML rather than by fmt, whose percent verbs the CSS would trip over.
const videoHTMLTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<style>
html, body { margin: 0; background: transparent; }
video { display: block; width: 100%; }
</style>
</head>
<body>
<video src="{video_url}" controls playsinline preload="metadata"></video>
</body>
</html>`
