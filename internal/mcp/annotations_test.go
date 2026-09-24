package mcp

import (
	"context"
	"testing"
)

func TestToolAnnotations(t *testing.T) {
	srv, err := New(Deps{Log: zapNop(), Resolver: newResolver(t)})
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	result, err := connectSession(t, srv).ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools() error: %v", err)
	}

	wantOpenWorld := map[string]bool{"inline": true, "since": false, "sleep": false, "time": false}

	if len(result.Tools) != len(wantOpenWorld) {
		t.Fatalf("tools = %d, want %d", len(result.Tools), len(wantOpenWorld))
	}

	for _, tool := range result.Tools {
		ann := tool.Annotations
		if ann == nil {
			t.Errorf("tool %s annotations = nil, want hints", tool.Name)

			continue
		}

		if !ann.ReadOnlyHint {
			t.Errorf("tool %s readOnlyHint = false, want true", tool.Name)
		}

		if ann.DestructiveHint == nil || *ann.DestructiveHint {
			t.Errorf("tool %s destructiveHint = %v, want explicit false", tool.Name, ann.DestructiveHint)
		}

		if !ann.IdempotentHint {
			t.Errorf("tool %s idempotentHint = false, want true", tool.Name)
		}

		openWorld, ok := wantOpenWorld[tool.Name]
		if !ok {
			t.Errorf("tool %s has no expected openWorldHint", tool.Name)

			continue
		}

		if ann.OpenWorldHint == nil || *ann.OpenWorldHint != openWorld {
			t.Errorf("tool %s openWorldHint = %v, want explicit %v", tool.Name, ann.OpenWorldHint, openWorld)
		}
	}
}
