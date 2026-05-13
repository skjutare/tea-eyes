// Package server constructs the tea-eyes MCP server and registers its tools.
package server

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"gitlab.com/skjutare/tea-eyes/internal/capture"
	capturedrv "gitlab.com/skjutare/tea-eyes/internal/driver"
	"gitlab.com/skjutare/tea-eyes/internal/keys"
	"gitlab.com/skjutare/tea-eyes/internal/render"
	"gitlab.com/skjutare/tea-eyes/internal/teatest"
	"gitlab.com/skjutare/tea-eyes/internal/tmux"
)

// Version is the server version reported during MCP initialize. Overridable
// from main via ldflags or a setter.
var Version = "0.1.0-dev"

// CaptureTextResult is the structured payload returned by tui_capture_text.
type CaptureTextResult struct {
	Text        string `json:"text"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	RawBytes    int    `json:"raw_bytes"`
	Mode        string `json:"mode"`
	TmuxSession string `json:"tmux_session,omitempty"`
}

// SessionAttachHintResult is the structured payload returned by
// tui_session_attach_hint.
type SessionAttachHintResult struct {
	Command string `json:"command"`
	Exists  bool   `json:"exists"`
}

// New builds a fully-configured MCP server with all tea-eyes tools registered,
// using the default render cache directory.
func New() *server.MCPServer {
	return NewWithRenderer(render.NewRenderer(""))
}

// NewWithRenderer is like New but allows callers (mainly tests) to supply a
// pre-configured Renderer — e.g. with a custom cache directory.
func NewWithRenderer(r *render.Renderer) *server.MCPServer {
	s := server.NewMCPServer(
		"tea-eyes",
		Version,
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)
	registerCaptureText(s)
	registerRenderImage(s, r)
	registerTestGolden(s, teatest.NewDriver(""))
	registerInspectModel(s, teatest.NewDriver(""))
	registerSessionAttachHint(s, tmux.New())
	return s
}

func registerCaptureText(s *server.MCPServer) {
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
		mcp.WithString("mode",
			mcp.Description(
				"Capture backend: \"pty\" (default, ephemeral pseudo-terminal) or "+
					"\"tmux\" (persistent session the user can attach to in another "+
					"terminal).",
			),
			mcp.DefaultString("pty"),
			mcp.Enum("pty", "tmux"),
		),
		mcp.WithString("tmux_session",
			mcp.Description(
				"tmux mode only. Existing session name to attach to (and create if "+
					"missing). Leave empty to create an ephemeral teaeyes-<rand> "+
					"session that is killed after the call unless tmux_persist=true.",
			),
		),
		mcp.WithBoolean("tmux_persist",
			mcp.Description(
				"tmux mode only. Keep the session alive after the call so the user "+
					"(or a follow-up call) can attach to it. Returns tmux_session "+
					"in the structured result.",
			),
			mcp.DefaultBool(false),
		),
	)
	s.AddTool(tool, makeCaptureTextHandler())
}

// captureTextInput is the validated, parsed form of a tui_capture_text call.
type captureTextInput struct {
	command     string
	args        []string
	keys        [][]byte
	width       int
	height      int
	settleMs    int
	stripANSI   bool
	cwd         string
	mode        capturedrv.Mode
	tmuxSession string
	tmuxPersist bool
}

func parseCaptureTextInput(args map[string]any) (captureTextInput, *mcp.CallToolResult) {
	in := captureTextInput{
		width:     intArg(args, "width", 80),
		height:    intArg(args, "height", 24),
		settleMs:  intArg(args, "settle_ms", 300),
		stripANSI: boolArg(args, "strip_ansi", true),
	}
	in.command, _ = args["command"].(string)
	if in.command == "" {
		return in, mcp.NewToolResultError("tui_capture_text: 'command' is required")
	}
	in.cwd, _ = args["cwd"].(string)

	strArgs, err := stringSlice(args, "args")
	if err != nil {
		return in, mcp.NewToolResultError("tui_capture_text: " + err.Error())
	}
	in.args = strArgs
	keySpecs, err := stringSlice(args, "keys")
	if err != nil {
		return in, mcp.NewToolResultError("tui_capture_text: " + err.Error())
	}

	modeStr, _ := args["mode"].(string)
	mode, mErr := capturedrv.ParseMode(modeStr)
	if mErr != nil {
		return in, mcp.NewToolResultError("tui_capture_text: " + mErr.Error())
	}
	in.mode = mode
	in.tmuxSession, _ = args["tmux_session"].(string)
	in.tmuxPersist = boolArg(args, "tmux_persist", false)
	if in.mode == capturedrv.ModePTY && (in.tmuxSession != "" || in.tmuxPersist) {
		return in, mcp.NewToolResultError(
			"tui_capture_text: tmux_session and tmux_persist require mode=\"tmux\"",
		)
	}

	in.keys = make([][]byte, 0, len(keySpecs))
	for i, spec := range keySpecs {
		b, perr := keys.Parse(spec)
		if perr != nil {
			return in, mcp.NewToolResultError(
				fmt.Sprintf("tui_capture_text: keys[%d]=%q: %v", i, spec, perr),
			)
		}
		in.keys = append(in.keys, b)
	}
	return in, nil
}

func makeCaptureTextHandler() server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		in, bad := parseCaptureTextInput(req.GetArguments())
		if bad != nil {
			return bad, nil
		}

		drv, err := capturedrv.New(in.mode)
		if err != nil {
			return mcp.NewToolResultError("tui_capture_text: " + err.Error()), nil
		}

		out, err := drv.Capture(ctx, capturedrv.CaptureOpts{
			Command:     in.command,
			Args:        in.args,
			Keys:        in.keys,
			Width:       in.width,
			Height:      in.height,
			SettleMs:    in.settleMs,
			Cwd:         in.cwd,
			SessionName: in.tmuxSession,
			Persist:     in.tmuxPersist,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("tui_capture_text: capture failed", err), nil
		}

		frame, err := capture.RenderFrame(out.Raw, in.width, in.height, in.stripANSI)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("tui_capture_text: render failed", err), nil
		}

		result := CaptureTextResult{
			Text:        frame.Text,
			Width:       frame.Width,
			Height:      frame.Height,
			RawBytes:    len(out.Raw),
			Mode:        string(in.mode),
			TmuxSession: out.Session,
		}
		return mcp.NewToolResultStructured(result, frame.Text), nil
	}
}

func registerSessionAttachHint(s *server.MCPServer, td *tmux.Driver) {
	tool := mcp.NewTool("tui_session_attach_hint",
		mcp.WithDescription(
			"Return the shell command the user can run to attach to a persistent "+
				"tmux session created by tea-eyes (typically from a tui_capture_text "+
				"call with mode=\"tmux\" and tmux_persist=true). Use this to tell the "+
				"user exactly how to watch Claude drive the TUI live.",
		),
		mcp.WithString("session_name",
			mcp.Description("Name of the tmux session to attach to."),
			mcp.Required(),
		),
	)
	s.AddTool(tool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		name, _ := args["session_name"].(string)
		if name == "" {
			return mcp.NewToolResultError("tui_session_attach_hint: 'session_name' is required"), nil
		}
		exists := false
		if _, lpErr := td.LookPath(); lpErr == nil {
			ok, _ := td.HasSession(ctx, name)
			exists = ok
		}
		cmd := "tmux attach -t " + name
		summary := cmd
		if !exists {
			summary = cmd + "  (session not currently present)"
		}
		return mcp.NewToolResultStructured(SessionAttachHintResult{
			Command: cmd,
			Exists:  exists,
		}, summary), nil
	})
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
		s, isStr := item.(string)
		if !isStr {
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
