package integration

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func skipIfNoTmux(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("tmux"); err != nil {
		t.Skip("tmux not installed; skipping tmux integration tests")
	}
}

func killSession(t *testing.T, name string) {
	t.Helper()
	if name == "" {
		return
	}
	_ = exec.Command("tmux", "kill-session", "-t", name).Run()
}

func captureViaTmux(
	t *testing.T,
	c *client.Client,
	ctx context.Context,
	args map[string]any,
) (text string, raw map[string]any) {
	t.Helper()
	args["mode"] = "tmux"
	return callCapture(t, c, ctx, args)
}

func TestTuiCaptureText_Tmux_Greeting(t *testing.T) {
	skipIfNoTmux(t)
	bin := helloTUI(t)
	c, ctx, cancel := newClient(t)
	defer cancel()

	text, raw := captureViaTmux(t, c, ctx, map[string]any{
		"command":   bin,
		"width":     80,
		"height":    24,
		"settle_ms": 400,
	})

	if !strings.Contains(text, "Hello, tea-eyes!") {
		t.Errorf("expected greeting under tmux mode, got:\n%s", text)
	}
	if mode, _ := raw["mode"].(string); mode != "tmux" {
		t.Errorf("expected mode=tmux in result, got %v", raw["mode"])
	}
	// Ephemeral session: tmux_session should be empty in the result.
	if s, _ := raw["tmux_session"].(string); s != "" {
		t.Errorf("ephemeral tmux session should not be returned, got %q", s)
		killSession(t, s)
	}
}

func TestTuiCaptureText_Tmux_PersistAndReuse(t *testing.T) {
	skipIfNoTmux(t)
	bin := helloTUI(t)
	c, ctx, cancel := newClient(t)
	defer cancel()

	_, raw := captureViaTmux(t, c, ctx, map[string]any{
		"command":      bin,
		"width":        80,
		"height":       24,
		"settle_ms":    400,
		"tmux_persist": true,
	})
	session, _ := raw["tmux_session"].(string)
	if session == "" {
		t.Fatal("tmux_persist=true should return a tmux_session in the result")
	}
	t.Cleanup(func() { killSession(t, session) })

	// Reuse the same session. The hello-tui binary will have exited by now;
	// re-run it inside the existing session and verify it still works.
	text2, raw2 := captureViaTmux(t, c, ctx, map[string]any{
		"command":      bin,
		"width":        80,
		"height":       24,
		"settle_ms":    400,
		"tmux_session": session,
		"tmux_persist": true,
	})
	if got, _ := raw2["tmux_session"].(string); got != session {
		t.Errorf("expected returned tmux_session=%q, got %q", session, got)
	}
	if !strings.Contains(text2, "Hello, tea-eyes!") {
		t.Errorf("expected greeting on session reuse, got:\n%s", text2)
	}
}

func TestTuiCaptureText_PTY_RejectsTmuxArgs(t *testing.T) {
	bin := helloTUI(t)
	c, ctx, cancel := newClient(t)
	defer cancel()

	req := mcp.CallToolRequest{}
	req.Params.Name = "tui_capture_text"
	req.Params.Arguments = map[string]any{
		"command":      bin,
		"tmux_persist": true,
	}
	res, err := c.CallTool(ctx, req)
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected error when tmux_* used with default pty mode")
	}
}

func TestTuiRenderImage_RejectsTmuxMode(t *testing.T) {
	bin := helloTUI(t)
	c, ctx, cancel := newClient(t)
	defer cancel()

	req := mcp.CallToolRequest{}
	req.Params.Name = "tui_render_image"
	req.Params.Arguments = map[string]any{
		"command": bin,
		"mode":    "tmux",
	}
	res, err := c.CallTool(ctx, req)
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected mode=tmux to be rejected by tui_render_image")
	}
	var msg strings.Builder
	for _, ct := range res.Content {
		if tc, ok := ct.(mcp.TextContent); ok {
			msg.WriteString(tc.Text)
		}
	}
	if !strings.Contains(msg.String(), "tmux") {
		t.Errorf("error should mention tmux limitation, got: %s", msg.String())
	}
}

func TestTuiSessionAttachHint(t *testing.T) {
	skipIfNoTmux(t)
	c, ctx, cancel := newClient(t)
	defer cancel()

	// Create a session via the MCP tool with persist=true so the hint reports
	// exists=true.
	bin := helloTUI(t)
	_, raw := captureViaTmux(t, c, ctx, map[string]any{
		"command":      bin,
		"width":        80,
		"height":       24,
		"settle_ms":    400,
		"tmux_persist": true,
	})
	session, _ := raw["tmux_session"].(string)
	if session == "" {
		t.Fatal("setup: expected persistent tmux session")
	}
	t.Cleanup(func() { killSession(t, session) })

	req := mcp.CallToolRequest{}
	req.Params.Name = "tui_session_attach_hint"
	req.Params.Arguments = map[string]any{"session_name": session}
	res, err := c.CallTool(ctx, req)
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if res.IsError {
		t.Fatalf("attach hint returned error: %+v", res.Content)
	}
	if res.StructuredContent == nil {
		t.Fatal("expected structured result")
	}
	b, _ := json.Marshal(res.StructuredContent)
	var hint struct {
		Command string `json:"command"`
		Exists  bool   `json:"exists"`
	}
	_ = json.Unmarshal(b, &hint)
	want := "tmux attach -t " + session
	if hint.Command != want {
		t.Errorf("expected command=%q, got %q", want, hint.Command)
	}
	if !hint.Exists {
		t.Errorf("expected exists=true for live session, got false")
	}
}

func TestTuiTestGolden_RejectsModeArg(t *testing.T) {
	c, ctx, cancel := newClient(t)
	defer cancel()

	req := mcp.CallToolRequest{}
	req.Params.Name = "tui_test_golden"
	req.Params.Arguments = map[string]any{
		"package_path": "../../examples/hello-tui",
		"golden_file":  "irrelevant.golden",
		"mode":         "tmux",
	}
	res, err := c.CallTool(ctx, req)
	if err != nil {
		t.Fatalf("call tool: %v", err)
	}
	if !res.IsError {
		t.Fatal("expected mode arg to be rejected by tui_test_golden")
	}
}
