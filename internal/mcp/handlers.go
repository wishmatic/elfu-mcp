package mcp

import (
	"time"

	"github.com/wishmatic/elfu-mcp/internal/resolve"
	"go.uber.org/zap"
)

type handlers struct {
	log      *zap.Logger
	resolver *resolve.Resolver
	location *time.Location
	now      func() time.Time
}
