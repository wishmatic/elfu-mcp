package mcp

import (
	"context"
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

func TestToolInputSchemaIsAnObject(t *testing.T) {
	srv, err := New(Deps{Log: zapNop(), Resolver: newResolver(t)})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	result, err := connectSession(t, srv).ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools() error: %v", err)
	}

	for _, tool := range result.Tools {
		schema, ok := tool.InputSchema.(map[string]any)
		if !ok {
			t.Fatalf("tool %s input schema = %#v, want an object schema", tool.Name, tool.InputSchema)
		}

		if schema["type"] != "object" {
			t.Errorf("tool %s input schema type = %v, want object", tool.Name, schema["type"])
		}
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
			want:       []string{"inline", "since", "sleep", "time"},
		},
		{
			name: "no resolver",
			want: []string{"since", "sleep", "time"},
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
