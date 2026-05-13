package integration

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func callTool(t *testing.T, c *client.Client, ctx context.Context, name string, args map[string]any) (raw map[string]any, isErr bool, errMsg string) {
	t.Helper()
	req := mcp.CallToolRequest{}
	req.Params.Name = name
	req.Params.Arguments = args
	res, err := c.CallTool(ctx, req)
	if err != nil {
		t.Fatalf("%s call: %v", name, err)
	}
	if res.IsError {
		for _, ct := range res.Content {
			if tc, ok := ct.(mcp.TextContent); ok {
				errMsg += tc.Text
			}
		}
		return nil, true, errMsg
	}
	if res.StructuredContent != nil {
		b, _ := json.Marshal(res.StructuredContent)
		_ = json.Unmarshal(b, &raw)
	}
	return raw, false, ""
}

func TestTuiTestGolden_CreateAndCompare(t *testing.T) {
	c, ctx, cancel := newClient(t)
	defer cancel()

	pkg, err := filepath.Abs("../../examples/hello-tui")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}

	tmp := t.TempDir()
	golden := filepath.Join(tmp, "hello.golden")

	// 1. First call: golden file does not exist — driver should create it.
	out, isErr, msg := callTool(t, c, ctx, "tui_test_golden", map[string]any{
		"package_path": pkg,
		"keys":         []any{"space"},
		"width":        40,
		"height":       8,
		"golden_file":  golden,
	})
	if isErr {
		t.Fatalf("tui_test_golden create returned error: %s", msg)
	}
	if match, _ := out["match"].(bool); !match {
		t.Fatalf("expected match=true on create, got %v", out)
	}
	if created, _ := out["created"].(bool); !created {
		t.Fatalf("expected created=true on first call, got %v", out)
	}
	if _, err := os.Stat(golden); err != nil {
		t.Fatalf("golden file should exist: %v", err)
	}

	// 2. Second call: compare against the just-written golden — should match.
	out, isErr, msg = callTool(t, c, ctx, "tui_test_golden", map[string]any{
		"package_path": pkg,
		"keys":         []any{"space"},
		"width":        40,
		"height":       8,
		"golden_file":  golden,
	})
	if isErr {
		t.Fatalf("tui_test_golden compare returned error: %s", msg)
	}
	if match, _ := out["match"].(bool); !match {
		t.Fatalf("expected match=true on second call, got %v", out)
	}
	if created, _ := out["created"].(bool); created {
		t.Fatalf("expected created=false on second call, got %v", out)
	}

	// 3. Mutate the input (different key sequence) — should mismatch with a diff.
	out, isErr, msg = callTool(t, c, ctx, "tui_test_golden", map[string]any{
		"package_path": pkg,
		"keys":         []any{"space", "space"},
		"width":        40,
		"height":       8,
		"golden_file":  golden,
	})
	if isErr {
		t.Fatalf("tui_test_golden mutated returned error: %s", msg)
	}
	if match, _ := out["match"].(bool); match {
		t.Fatalf("expected match=false after mutating keys, got %v", out)
	}
	diff, _ := out["diff"].(string)
	if !strings.Contains(diff, "Counter:") {
		t.Fatalf("expected diff to mention the counter line, got:\n%s", diff)
	}
}

func TestTuiInspectModel(t *testing.T) {
	c, ctx, cancel := newClient(t)
	defer cancel()

	pkg, err := filepath.Abs("../../examples/hello-tui")
	if err != nil {
		t.Fatalf("abs: %v", err)
	}

	out, isErr, msg := callTool(t, c, ctx, "tui_inspect_model", map[string]any{
		"package_path": pkg,
		"keys":         []any{"space"},
		"width":        40,
		"height":       8,
	})
	if isErr {
		t.Fatalf("tui_inspect_model returned error: %s", msg)
	}
	mj, _ := out["model_json"].(string)
	if mj == "" || mj == "null" {
		t.Fatalf("expected non-empty model_json, got %q", mj)
	}
	if !strings.HasPrefix(strings.TrimSpace(mj), "{") {
		t.Fatalf("expected model_json to be a JSON object, got: %s", mj)
	}
	if !strings.Contains(mj, "Counter") || !strings.Contains(mj, "ShowCounter") {
		t.Fatalf("expected model_json to include exported fields Counter and ShowCounter, got: %s", mj)
	}
	vt, _ := out["view_text"].(string)
	if !strings.Contains(vt, "Hello, tea-eyes!") {
		t.Fatalf("expected view_text to contain greeting, got:\n%s", vt)
	}
}
