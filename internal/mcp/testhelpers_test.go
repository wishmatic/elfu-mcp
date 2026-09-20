package mcp

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/wishmatic/elfu-mcp/internal/resolve"
	"go.uber.org/zap"
)

const testImageSize = 4

func zapNop() *zap.Logger {
	return zap.NewNop()
}

func newResolver(t *testing.T) *resolve.Resolver {
	t.Helper()

	resolver, err := resolve.New(nil, "", nil)
	if err != nil {
		t.Fatalf("resolve.New() error: %v", err)
	}

	return resolver
}

func testImagePNG(t *testing.T) []byte {
	t.Helper()

	img := image.NewNRGBA(image.Rect(0, 0, testImageSize, testImageSize))

	for y := 0; y < testImageSize; y++ {
		for x := 0; x < testImageSize; x++ {
			img.SetNRGBA(x, y, color.NRGBA{R: 180, G: 40, B: 10, A: 255})
		}
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("encode png: %v", err)
	}

	return buf.Bytes()
}
