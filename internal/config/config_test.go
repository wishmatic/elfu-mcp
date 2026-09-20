package config

import (
	"os"
	"path/filepath"
	"testing"
)

func unsetEnv(t *testing.T, key string) {
	t.Helper()

	prev, had := os.LookupEnv(key)

	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("unset %s: %v", key, err)
	}

	t.Cleanup(func() {
		if had {
			_ = os.Setenv(key, prev)

			return
		}

		_ = os.Unsetenv(key)
	})
}

func TestLoadDefaults(t *testing.T) {
	for _, key := range []string{"HOST", "PORT", "LOG_LEVEL", "PUBLIC_HOST", "FILES_DIR"} {
		unsetEnv(t, key)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Host != "0.0.0.0" {
		t.Errorf("Host = %q, want 0.0.0.0", cfg.Host)
	}

	if cfg.Port != 8080 {
		t.Errorf("Port = %d, want 8080", cfg.Port)
	}

	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want info", cfg.LogLevel)
	}

	if cfg.PublicHost != "" {
		t.Errorf("PublicHost = %q, want empty by default", cfg.PublicHost)
	}

	if cfg.FilesDir != "files" {
		t.Errorf("FilesDir = %q, want files by default", cfg.FilesDir)
	}

	if got := cfg.Addr(); got != "0.0.0.0:8080" {
		t.Errorf("Addr() = %q, want 0.0.0.0:8080", got)
	}
}

func TestLoadFiles(t *testing.T) {
	t.Setenv("PUBLIC_HOST", "https://elfu.example.com")
	t.Setenv("FILES_DIR", filepath.Join("data", "files"))

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.PublicHost != "https://elfu.example.com" {
		t.Errorf("PublicHost = %q, want https://elfu.example.com", cfg.PublicHost)
	}

	if cfg.FilesDir != filepath.Join("data", "files") {
		t.Errorf("FilesDir = %q, want the configured path", cfg.FilesDir)
	}
}

func TestPublicBase(t *testing.T) {
	tests := []struct {
		name       string
		publicHost string
		want       string
		wantErr    bool
	}{
		{name: "unset", publicHost: "", wantErr: false},
		{name: "valid", publicHost: "https://elfu.example.com", want: "https://elfu.example.com"},
		{name: "valid with port", publicHost: "http://192.168.1.10:8080", want: "http://192.168.1.10:8080"},
		{name: "trailing slash trimmed", publicHost: "https://elfu.example.com/", want: "https://elfu.example.com"},
		{name: "missing scheme", publicHost: "elfu.example.com", wantErr: true},
		{name: "non-http scheme", publicHost: "ftp://elfu.example.com", wantErr: true},
		{name: "empty host", publicHost: "https://", wantErr: true},
		{name: "path component", publicHost: "https://elfu.example.com/elfu", wantErr: true},
		{name: "query component", publicHost: "https://elfu.example.com?x=1", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := Config{PublicHost: tt.publicHost}

			base, err := cfg.PublicBase()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("PublicBase() error = nil, want an error")
				}

				return
			}

			if err != nil {
				t.Fatalf("PublicBase() error: %v", err)
			}

			if tt.publicHost == "" {
				if base != nil {
					t.Fatalf("PublicBase() = %v, want nil when PUBLIC_HOST is unset", base)
				}

				return
			}

			if base == nil || base.Host == "" {
				t.Fatalf("PublicBase() = %v, want a URL with a host", base)
			}

			if got := base.String(); got != tt.want {
				t.Errorf("PublicBase() = %q, want %q", got, tt.want)
			}
		})
	}
}
