package server

import (
	"bytes"
	"context"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"go.uber.org/zap"
)

type bearerRoundTripper struct {
	token string
	base  http.RoundTripper
}

func (t bearerRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.Header.Set("Authorization", "Bearer "+t.token)

	return t.base.RoundTrip(clone)
}

func testPNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))

	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 10, G: 90, B: 200, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}

	return buf.Bytes()
}

func newAPI(t *testing.T, srv *Server) *httptest.Server {
	t.Helper()

	api := httptest.NewServer(srv.router)
	t.Cleanup(api.Close)

	return api
}

func newImageServer(t *testing.T) *httptest.Server {
	t.Helper()

	images := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(testPNG(t))
	}))

	t.Cleanup(images.Close)

	return images
}

func newMCPSession(t *testing.T, api *httptest.Server) *mcp.ClientSession {
	t.Helper()

	session, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "v0"}, nil).Connect(
		context.Background(),
		&mcp.StreamableClientTransport{
			Endpoint: api.URL + "/mcp",
			HTTPClient: &http.Client{Transport: bearerRoundTripper{
				token: testAPIKey,
				base:  http.DefaultTransport,
			}},
			DisableStandaloneSSE: true,
		},
		nil,
	)
	if err != nil {
		t.Fatalf("Connect() error: %v", err)
	}

	t.Cleanup(func() { _ = session.Close() })

	return session
}

func callInline(t *testing.T, session *mcp.ClientSession, imageURL string) *mcp.CallToolResult {
	t.Helper()

	result, err := session.CallTool(context.Background(), &mcp.CallToolParams{
		Name:      "inline",
		Arguments: map[string]any{"image_url": imageURL},
	})
	if err != nil {
		t.Fatalf("CallTool(inline) error: %v", err)
	}

	if result.IsError {
		t.Fatalf("CallTool(inline) tool error: %+v", result.Content)
	}

	return result
}

func TestMCPListsTools(t *testing.T) {
	api := newAPI(t, newTestServer(t, zap.NewNop()))
	session := newMCPSession(t, api)

	result, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools() error: %v", err)
	}

	names := make([]string, 0, len(result.Tools))
	for _, tool := range result.Tools {
		names = append(names, tool.Name)
	}

	if want := []string{"inline", "since", "time"}; !slices.Equal(names, want) {
		t.Errorf("tools = %v, want %v", names, want)
	}
}

func TestMCPInlineOverHTTP(t *testing.T) {
	images := newImageServer(t)
	api := newAPI(t, newTestServer(t, zap.NewNop()))
	session := newMCPSession(t, api)

	result := callInline(t, session, images.URL+"/cat.png")

	if len(result.Content) != 2 {
		t.Fatalf("content = %d, want a caption and one image", len(result.Content))
	}

	caption, ok := result.Content[0].(*mcp.TextContent)
	if !ok || !strings.Contains(caption.Text, images.URL) {
		t.Fatalf("content[0] = %#v, want a caption naming the URL", result.Content[0])
	}

	img, ok := result.Content[1].(*mcp.ImageContent)
	if !ok {
		t.Fatalf("content[1] = %#v, want an image block", result.Content[1])
	}

	if img.MIMEType != "image/webp" {
		t.Errorf("mime type = %q, want image/webp", img.MIMEType)
	}

	if _, format, err := image.Decode(bytes.NewReader(img.Data)); err != nil || format != "webp" {
		t.Errorf("decode inline image = %q, %v, want webp", format, err)
	}
}

func TestMCPInlineReadsOwnPublicHostFromStore(t *testing.T) {
	cfg := testConfig(t)
	cfg.PublicHost = "http://files.internal:9"

	srv, err := New(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Cleanup(func() { _ = srv.Shutdown(context.Background()) })

	stored, err := srv.files.UploadFile(context.Background(), testPNG(t), "image/png")
	if err != nil {
		t.Fatalf("UploadFile() error: %v", err)
	}

	api := newAPI(t, srv)
	session := newMCPSession(t, api)

	result := callInline(t, session, stored)

	if len(result.Content) != 2 {
		t.Fatalf("content = %d, want a caption and one image", len(result.Content))
	}
}

func TestMCPInlineReadsMappedDirectory(t *testing.T) {
	dir := t.TempDir()

	if err := os.WriteFile(filepath.Join(dir, "pasted.png"), testPNG(t), 0o640); err != nil {
		t.Fatalf("WriteFile() error: %v", err)
	}

	cfg := testConfig(t)
	cfg.ImageURLMap = "https://chat.example.com/images/=" + dir

	srv, err := New(cfg, zap.NewNop())
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Cleanup(func() { _ = srv.Shutdown(context.Background()) })

	api := newAPI(t, srv)
	session := newMCPSession(t, api)

	result := callInline(t, session, "https://chat.example.com/images/pasted.png")

	img, ok := result.Content[1].(*mcp.ImageContent)
	if !ok {
		t.Fatalf("content[1] = %#v, want an image block", result.Content[1])
	}

	if img.MIMEType != "image/webp" {
		t.Errorf("mime type = %q, want image/webp", img.MIMEType)
	}
}

func TestMCPRejectsUnauthenticatedRequest(t *testing.T) {
	api := newAPI(t, newTestServer(t, zap.NewNop()))

	resp, err := http.Post(api.URL+"/mcp", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatalf("Post() error: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", resp.StatusCode)
	}
}

func TestServesStoredFileOverHTTP(t *testing.T) {
	srv := newTestServer(t, zap.NewNop())
	api := newAPI(t, srv)

	stored, err := srv.files.UploadFile(context.Background(), testPNG(t), "image/png")
	if err != nil {
		t.Fatalf("UploadFile() error: %v", err)
	}

	parsed, err := url.Parse(stored)
	if err != nil {
		t.Fatalf("url.Parse() error: %v", err)
	}

	resp, err := http.Get(api.URL + parsed.Path)
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	if ct := resp.Header.Get("Content-Type"); ct != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", ct)
	}

	if cache := resp.Header.Get("Cache-Control"); !strings.Contains(cache, "immutable") {
		t.Errorf("Cache-Control = %q, want immutable", cache)
	}
}
