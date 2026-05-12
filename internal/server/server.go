// Package server constructs the tea-eyes MCP server and registers its tools.
package server

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"gitlab.com/skjutare/tea-eyes/internal/capture"
	"gitlab.com/skjutare/tea-eyes/internal/keys"
	"gitlab.com/skjutare/tea-eyes/internal/pty"
)

// Version is the server version reported during MCP initialize. Overridable
// from main via ldflags or a setter.
var Version = "0.1.0-dev"

// CaptureTextResult is the structured payload returned by tui_capture_text.
type CaptureTextResult struct {
	Text     string `json:"text"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	RawBytes int    `json:"raw_bytes"`
}

// New builds a fully-configured MCP server with all tea-eyes tools registered.
func New() *server.MCPServer {
	s := server.NewMCPServer(
		"tea-eyes",
		Version,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)
	registerCaptureText(s, pty.New())
	return s
}

func registerCaptureText(s *server.MCPServer, driver *pty.Driver) {
	tool := mcp.NewTool("tui_capture_text",
		mcp.WithDescription(
			"Spawn a TUI command under a pseudo-terminal, optionally send a "+
				"sequence of keystrokes, and return the rendered text grid. "+
				"Use this to see what a Bubble Tea (or any) TUI displays.",
		),
		mcp.WithString("command",
			mcp.Description("Path to the TUI binary (or command on PATH) to run."),
			mcp.Required(),
		),
		mcp.WithArray("args",
			mcp.Description("Arguments to pass to the command."),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithArray("keys",
			mcp.Description(
				"Sequence of key inputs to send after the initial render. "+
					"Each entry is a key spec like \"enter\", \"ctrl+c\", "+
					"\"shift+tab\", or a literal string like \"hello\".",
			),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithNumber("width",
			mcp.Description("Terminal width in columns."),
			mcp.DefaultNumber(80),
		),
		mcp.WithNumber("height",
			mcp.Description("Terminal height in rows."),
			mcp.DefaultNumber(24),
		),
		mcp.WithNumber("settle_ms",
			mcp.Description("Milliseconds to wait after spawn and between key sends."),
			mcp.DefaultNumber(300),
		),
		mcp.WithBoolean("strip_ansi",
			mcp.Description("If true, return a clean text grid; if false, return raw bytes with SGR escapes."),
			mcp.DefaultBool(true),
		),
		mcp.WithString("cwd",
			mcp.Description("Working directory for the spawned command. Defaults to the server's cwd."),
		),
	)
	s.AddTool(tool, makeCaptureTextHandler(driver))
}

func makeCaptureTextHandler(driver *pty.Driver) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		command, _ := args["command"].(string)
		if command == "" {
			return mcp.NewToolResultError("tui_capture_text: 'command' is required"), nil
		}

		strArgs, err := stringSlice(args, "args")
		if err != nil {
			return mcp.NewToolResultError("tui_capture_text: " + err.Error()), nil
		}
		keySpecs, err := stringSlice(args, "keys")
		if err != nil {
			return mcp.NewToolResultError("tui_capture_text: " + err.Error()), nil
		}

		width := intArg(args, "width", 80)
		height := intArg(args, "height", 24)
		settleMs := intArg(args, "settle_ms", 300)
		stripANSI := boolArg(args, "strip_ansi", true)
		cwd, _ := args["cwd"].(string)

		// Parse each key spec individually so per-entry errors are precise.
		keyBytes := make([][]byte, 0, len(keySpecs))
		for i, spec := range keySpecs {
			b, perr := keys.Parse(spec)
			if perr != nil {
				return mcp.NewToolResultError(
					fmt.Sprintf("tui_capture_text: keys[%d]=%q: %v", i, spec, perr),
				), nil
			}
			keyBytes = append(keyBytes, b)
		}

		raw, err := driver.Capture(ctx, pty.SpawnOpts{
			Command:  command,
			Args:     strArgs,
			Width:    width,
			Height:   height,
			SettleMs: settleMs,
			Cwd:      cwd,
		}, keyBytes)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("tui_capture_text: capture failed", err), nil
		}

		frame, err := capture.RenderFrame(raw, width, height, stripANSI)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("tui_capture_text: render failed", err), nil
		}

		result := CaptureTextResult{
			Text:     frame.Text,
			Width:    frame.Width,
			Height:   frame.Height,
			RawBytes: len(raw),
		}
		return mcp.NewToolResultStructured(result, frame.Text), nil
	}
}

func stringSlice(args map[string]any, key string) ([]string, error) {
	v, ok := args[key]
	if !ok || v == nil {
		return nil, nil
	}
	arr, ok := v.([]any)
	if !ok {
		return nil, fmt.Errorf("%s must be an array of strings", key)
	}
	out := make([]string, 0, len(arr))
	for i, item := range arr {
		s, ok := item.(string)
		if !ok {
			return nil, fmt.Errorf("%s[%d] must be a string", key, i)
		}
		out = append(out, s)
	}
	return out, nil
}

func intArg(args map[string]any, key string, def int) int {
	v, ok := args[key]
	if !ok || v == nil {
		return def
	}
	switch n := v.(type) {
	case float64:
		return int(n)
	case int:
		return n
	case int64:
		return int(n)
	default:
		return def
	}
}

func boolArg(args map[string]any, key string, def bool) bool {
	v, ok := args[key]
	if !ok || v == nil {
		return def
	}
	b, ok := v.(bool)
	if !ok {
		return def
	}
	return b
}

