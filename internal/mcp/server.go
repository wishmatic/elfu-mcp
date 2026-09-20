package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/wishmatic/elfu-mcp/internal/resolve"
	"go.uber.org/zap"
)

type Deps struct {
	Log      *zap.Logger
	Resolver *resolve.Resolver
}

func New(deps Deps) (*mcp.Server, error) {
	srv := mcp.NewServer(&mcp.Implementation{
		Name:    "elfu-mcp",
		Version: "0.1.0",
	}, nil)

	registerTools(srv, buildHandlers(deps))

	return srv, nil
}

func buildHandlers(deps Deps) *handlers {
	return &handlers{
		log:      deps.Log,
		resolver: deps.Resolver,
	}
}

func registerTools(srv *mcp.Server, h *handlers) {
	if h.resolver != nil {
		registerInline(srv, h)
	}
}
