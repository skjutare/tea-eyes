package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"

	"gitlab.com/skjutare/tea-eyes/internal/server"
)

var helloBin string

func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "tea-eyes-int-*")
	if err != nil {
		panic(err)
	}
	defer os.RemoveAll(dir)

	helloBin = filepath.Join(dir, "hello-tui")
	cmd := exec.Command("go", "build", "-o", helloBin, "../../examples/hello-tui")
	if b, err := cmd.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "go build hello-tui failed: %v\n%s", err, b)
		os.Exit(1)
	}

	os.Exit(m.Run())
}

func helloTUI(t *testing.T) string {
	t.Helper()
	return helloBin
}

func newClient(t *testing.T) (*client.Client, context.Context, context.CancelFunc) {
	t.Helper()
	s := server.New()
	c, err := client.NewInProcessClient(s)
	if err != nil {
		t.Fatalf("in-process client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	t.Cleanup(func() { _ = c.Close() })

	if err := c.Start(ctx); err != nil {
		cancel()
		t.Fatalf("client start: %v", err)
	}
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{Name: "tea-eyes-test", Version: "0"}
	if _, err := c.Initialize(ctx, initReq); err != nil {
		cancel()
		t.Fatalf("initialize: %v", err)
	}
	return c, ctx, cancel
}

func callCapture(
	t *testing.T,
	c *client.Client,
	ctx context.Context,
	args map[string]any,
) (text string, raw map[string]any) {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = "tui_capture_text"
	req.Params.Arguments = args
	res, err := c.CallTool(ctx, req)
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if res.IsError {
		var msg string
		var msgSb80 strings.Builder
		for _, ct := range res.Content {
			if tc, ok := ct.(mcp.TextContent); ok {
				msgSb80.WriteString(tc.Text)
			}
		}
		msg += msgSb80.String()
		t.Fatalf("tool returned error: %s", msg)
	}
	if res.StructuredContent != nil {
		b, _ := json.Marshal(res.StructuredContent)
		_ = json.Unmarshal(b, &raw)
		if v, ok := raw["text"].(string); ok {
			text = v
		}
		return text, raw
	}
	var textSb95 strings.Builder
	for _, ct := range res.Content {
		if tc, ok := ct.(mcp.TextContent); ok {
			textSb95.WriteString(tc.Text)
		}
	}
	text += textSb95.String()
	return text, raw
}

func TestTuiCaptureText_Greeting(t *testing.T) {
	bin := helloTUI(t)
	c, ctx, cancel := newClient(t)
	defer cancel()

	text, raw := callCapture(t, c, ctx, map[string]any{
		"command":   bin,
		"width":     40,
		"height":    8,
		"settle_ms": 250,
	})

	if !strings.Contains(text, "Hello, tea-eyes!") {
		t.Errorf("expected greeting in output, got:\n%s", text)
	}
	if w, _ := raw["width"].(float64); int(w) != 40 {
		t.Errorf("width round-trip mismatch: %v", raw["width"])
	}
	if rb, _ := raw["raw_bytes"].(float64); rb <= 0 {
		t.Errorf("raw_bytes should be > 0, got %v", raw["raw_bytes"])
	}
}

func TestTuiCaptureText_Counter(t *testing.T) {
	bin := helloTUI(t)
	c, ctx, cancel := newClient(t)
	defer cancel()

	text, _ := callCapture(t, c, ctx, map[string]any{
		"command":   bin,
		"keys":      []any{"space"},
		"width":     40,
		"height":    8,
		"settle_ms": 250,
	})
	if !strings.Contains(text, "Counter: 1") {
		t.Errorf("expected counter after space, got:\n%s", text)
	}
}

func TestTuiCaptureText_Quit(t *testing.T) {
	bin := helloTUI(t)
	c, ctx, cancel := newClient(t)
	defer cancel()

	// 'q' makes the program exit gracefully — capture should return whatever
	// was rendered up to that point without erroring.
	text, _ := callCapture(t, c, ctx, map[string]any{
		"command":   bin,
		"keys":      []any{"q"},
		"width":     40,
		"height":    8,
		"settle_ms": 250,
	})
	if !strings.Contains(text, "Hello, tea-eyes!") {
		t.Errorf("expected greeting to have rendered before quit, got:\n%s", text)
	}
}

func TestTuiCaptureText_UnknownCommand(t *testing.T) {
	c, ctx, cancel := newClient(t)
	defer cancel()

	req := mcp.CallToolRequest{}
	req.Params.Name = "tui_capture_text"
	req.Params.Arguments = map[string]any{
		"command": "this-binary-does-not-exist-tea-eyes",
	}
	res, err := c.CallTool(ctx, req)
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected error result, got success")
	}
	var msg string
	var msgSb179 strings.Builder
	for _, ct := range res.Content {
		if tc, ok := ct.(mcp.TextContent); ok {
			msgSb179.WriteString(tc.Text)
		}
	}
	msg += msgSb179.String()
	if !strings.Contains(msg, "command not found") {
		t.Errorf("error should mention 'command not found', got: %s", msg)
	}
}

func TestTuiCaptureText_BadKey(t *testing.T) {
	bin := helloTUI(t)
	c, ctx, cancel := newClient(t)
	defer cancel()

	req := mcp.CallToolRequest{}
	req.Params.Name = "tui_capture_text"
	req.Params.Arguments = map[string]any{
		"command": bin,
		"keys":    []any{"ctrl+meow"},
	}
	res, err := c.CallTool(ctx, req)
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error for bad key")
	}
	var msg string
	var msgSb208 strings.Builder
	for _, ct := range res.Content {
		if tc, ok := ct.(mcp.TextContent); ok {
			msgSb208.WriteString(tc.Text)
		}
	}
	msg += msgSb208.String()
	if !strings.Contains(msg, "ctrl+meow") {
		t.Errorf("error should mention offending key, got: %s", msg)
	}
}
