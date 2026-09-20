package mcp

import (
	"slices"
	"testing"
)

func TestNewRegistersTools(t *testing.T) {
	srv, err := New(Deps{Log: zapNop(), Resolver: newResolver(t)})
	if err != nil {
		t.Fatalf("New() unexpected error: %v", err)
	}

	if srv == nil {
		t.Fatal("New() returned nil server")
	}
}

func TestToolRegistration(t *testing.T) {
	tests := []struct {
		name       string
		configured bool
		want       []string
	}{
		{
			name:       "resolver configured",
			configured: true,
			want:       []string{"inline"},
		},
		{
			name: "no resolver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			deps := Deps{Log: zapNop()}
			if tt.configured {
				deps.Resolver = newResolver(t)
			}

			srv, err := New(deps)
			if err != nil {
				t.Fatalf("New() error: %v", err)
			}

			got := toolNames(t, srv)
			want := slices.Clone(tt.want)

			slices.Sort(got)
			slices.Sort(want)

			if !slices.Equal(got, want) {
				t.Errorf("tools = %v, want %v", got, want)
			}
		})
	}
}
