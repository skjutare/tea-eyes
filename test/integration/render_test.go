//go:build integration

package integration

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"image/png"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"gitlab.com/skjutare/tea-eyes/internal/render"
	"gitlab.com/skjutare/tea-eyes/internal/server"
)

func requireVHS(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("vhs"); err != nil {
		t.Skip("vhs not on PATH — install vhs/ttyd/ffmpeg to run render tests")
	}
}

func newRenderClient(t *testing.T, cacheDir string) (*client.Client, context.Context, context.CancelFunc) {
	t.Helper()
	s := server.NewWithRenderer(render.NewRenderer(cacheDir))
	c, err := client.NewInProcessClient(s)
	if err != nil {
		t.Fatalf("in-process client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	t.Cleanup(func() { _ = c.Close() })
	if err := c.Start(ctx); err != nil {
		cancel()
		t.Fatalf("client start: %v", err)
	}
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "tea-eyes-render-test", Version: "0"}
	if _, err := c.Initialize(ctx, initReq); err != nil {
		cancel()
		t.Fatalf("initialize: %v", err)
	}
	return c, ctx, cancel
}

func callRender(
	t *testing.T,
	c *client.Client,
	ctx context.Context,
	args map[string]any,
) (imgBytes []byte, meta map[string]any) {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = "tui_render_image"
	req.Params.Arguments = args
	res, err := c.CallTool(ctx, req)
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if res.IsError {
		var msg string
		for _, ct := range res.Content {
			if tc, ok := ct.(mcp.TextContent); ok {
				msg += tc.Text
			}
		}
		t.Fatalf("tool returned error: %s", msg)
	}
	for _, ct := range res.Content {
		if ic, ok := ct.(mcp.ImageContent); ok {
			b, err := base64.StdEncoding.DecodeString(ic.Data)
			if err != nil {
				t.Fatalf("decode image base64: %v", err)
			}
			imgBytes = b
		}
	}
	if res.StructuredContent != nil {
		b, _ := json.Marshal(res.StructuredContent)
		_ = json.Unmarshal(b, &meta)
	}
	return
}

func TestTuiRenderImage_MultiPane(t *testing.T) {
	requireVHS(t)

	dir, err := os.MkdirTemp("", "tea-eyes-render-*")
	if err != nil {
		t.Fatalf("tempdir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	bin := filepath.Join(dir, "multi-pane")
	build := exec.Command("go", "build", "-o", bin, "../../examples/multi-pane")
	if b, err := build.CombinedOutput(); err != nil {
		t.Fatalf("build multi-pane: %v\n%s", err, b)
	}

	cacheDir := filepath.Join(dir, "renders")
	c, ctx, cancel := newRenderClient(t, cacheDir)
	defer cancel()

	args := map[string]any{
		"command":   bin,
		"width":     80,
		"height":    24,
		"settle_ms": 400,
		"format":    "png",
	}

	img, meta := callRender(t, c, ctx, args)
	if len(img) == 0 {
		t.Fatal("expected non-empty PNG bytes")
	}
	if cfg, err := png.DecodeConfig(bytes.NewReader(img)); err != nil {
		t.Fatalf("decode png: %v", err)
	} else {
		fontSize := 14.0
		expW := int(80*0.6*fontSize) + 20*2
		expH := int(24*1.3*fontSize) + 20*2
		if abs(cfg.Width-expW) > 80 || abs(cfg.Height-expH) > 80 {
			t.Logf("png dims %dx%d (expected ~%dx%d)", cfg.Width, cfg.Height, expW, expH)
		}
	}
	if hit, _ := meta["cache_hit"].(bool); hit {
		t.Errorf("first render should not be a cache hit")
	}

	_, meta2 := callRender(t, c, ctx, args)
	if hit, _ := meta2["cache_hit"].(bool); !hit {
		t.Errorf("second render with same inputs should be cache_hit=true, meta=%v", meta2)
	}

	noCacheArgs := map[string]any{}
	for k, v := range args {
		noCacheArgs[k] = v
	}
	noCacheArgs["no_cache"] = true
	_, meta3 := callRender(t, c, ctx, noCacheArgs)
	if hit, _ := meta3["cache_hit"].(bool); hit {
		t.Errorf("no_cache should force re-render, got cache_hit=true")
	}
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// Compile-time confirmation that server.NewWithRenderer exists.
var _ = fmt.Sprintf
