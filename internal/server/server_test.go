package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wishmatic/elfu-mcp/internal/auth"
	"github.com/wishmatic/elfu-mcp/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

const testAPIKey = "server-key"

func testConfig(t *testing.T) config.Config {
	t.Helper()

	return config.Config{
		APIKey:     testAPIKey,
		PublicHost: "https://elfu.example.com",
		FilesDir:   filepath.Join(t.TempDir(), "files"),
	}
}

func newTestServer(t *testing.T, log *zap.Logger) *Server {
	t.Helper()

	srv, err := New(testConfig(t), log)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Cleanup(func() { _ = srv.Shutdown(context.Background()) })

	return srv
}

func TestShutdown(t *testing.T) {
	srv := newTestServer(t, zap.NewNop())

	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("Shutdown() error: %v", err)
	}
}

func TestNewRequiresAPIKey(t *testing.T) {
	cfg := testConfig(t)
	cfg.APIKey = ""

	_, err := New(cfg, zap.NewNop())
	if err == nil {
		t.Fatal("New() error = nil, want an error")
	}

	if !strings.Contains(err.Error(), "API_KEY") {
		t.Errorf("error = %q, want it to name API_KEY", err.Error())
	}
}

func TestNewRequiresPublicHost(t *testing.T) {
	tests := []struct {
		name       string
		publicHost string
	}{
		{name: "missing", publicHost: ""},
		{name: "no scheme", publicHost: "elfu.example.com"},
		{name: "with path", publicHost: "https://elfu.example.com/elfu"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := testConfig(t)
			cfg.PublicHost = tt.publicHost

			_, err := New(cfg, zap.NewNop())
			if err == nil {
				t.Fatal("New() error = nil, want an error")
			}

			if !strings.Contains(err.Error(), "PUBLIC_HOST") {
				t.Errorf("error = %q, want it to name PUBLIC_HOST", err.Error())
			}
		})
	}
}

func TestNewRequiresFilesDir(t *testing.T) {
	cfg := testConfig(t)
	cfg.FilesDir = ""

	_, err := New(cfg, zap.NewNop())
	if err == nil {
		t.Fatal("New() error = nil, want an error")
	}

	if !strings.Contains(err.Error(), "FILES_DIR") {
		t.Errorf("error = %q, want it to name FILES_DIR", err.Error())
	}
}

func TestNewLogsLocalFilesWarning(t *testing.T) {
	core, logs := observer.New(zapcore.DebugLevel)

	cfg := testConfig(t)

	srv, err := New(cfg, zap.New(core))
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	t.Cleanup(func() { _ = srv.Shutdown(context.Background()) })

	var (
		enabled bool
		warned  bool
	)

	for _, entry := range logs.All() {
		if entry.Message == "local files enabled" && entry.ContextMap()["dir"] == cfg.FilesDir {
			enabled = true
		}

		if entry.Level == zapcore.WarnLevel && strings.Contains(entry.Message, "readable by anyone") {
			warned = true
		}
	}

	if !enabled {
		t.Error("no \"local files enabled\" log entry with the configured directory")
	}

	if !warned {
		t.Error("no warning that stored files are readable by anyone with the URL")
	}
}

func TestHealthz(t *testing.T) {
	srv := newTestServer(t, zap.NewNop())

	rec := httptest.NewRecorder()
	srv.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if rec.Body.String() != "ok" {
		t.Errorf("body = %q, want ok", rec.Body.String())
	}
}

func TestServesStoredFile(t *testing.T) {
	cfg := testConfig(t)
	srv := newTestServer(t, zap.NewNop())

	url, err := srv.files.UploadFile(context.Background(), []byte("png-bytes"), "image/png")
	if err != nil {
		t.Fatalf("UploadFile() error: %v", err)
	}

	if !strings.HasPrefix(url, cfg.PublicHost+"/i/") {
		t.Fatalf("url = %q, want a %s/i/ prefix", url, cfg.PublicHost)
	}

	rec := httptest.NewRecorder()
	srv.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}

	if rec.Body.String() != "png-bytes" {
		t.Errorf("body = %q, want png-bytes", rec.Body.String())
	}

	if ct := rec.Header().Get("Content-Type"); ct != "image/png" {
		t.Errorf("Content-Type = %q, want image/png", ct)
	}
}

func TestServesStoredFileNotFound(t *testing.T) {
	srv := newTestServer(t, zap.NewNop())

	url := "https://elfu.example.com/i/2026-09/00000000-0000-0000-0000-000000000000.png"

	rec := httptest.NewRecorder()
	srv.router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, url, nil))

	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", rec.Code)
	}
}

func TestMCPRequiresAPIKey(t *testing.T) {
	srv := newTestServer(t, zap.NewNop())

	rec := httptest.NewRecorder()
	srv.router.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader("{}")))

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401 without a bearer token", rec.Code)
	}

	rec = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/mcp", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer "+testAPIKey)
	srv.router.ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized {
		t.Error("status = 401 with the configured API key, want the request to reach the MCP handler")
	}
}

func TestAuthErrorIsSentinel(t *testing.T) {
	cfg := testConfig(t)
	cfg.APIKey = ""

	_, err := New(cfg, zap.NewNop())
	if !errors.Is(err, auth.ErrNoAPIKey) {
		t.Errorf("New() error = %v, want auth.ErrNoAPIKey", err)
	}
}
