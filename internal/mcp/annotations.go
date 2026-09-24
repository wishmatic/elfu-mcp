package mcp

import "github.com/modelcontextprotocol/go-sdk/mcp"

// toolAnnotations provide the same annotations to all tools in this MCP server, since they are all the same except
// potentially openWorld.
//
// If in future they diverge more, this will need to catch
func toolAnnotations(openWorld bool) *mcp.ToolAnnotations {
	return &mcp.ToolAnnotations{
		ReadOnlyHint:    true,
		DestructiveHint: new(false),
		IdempotentHint:  true,
		OpenWorldHint:   new(openWorld),
	}
}
