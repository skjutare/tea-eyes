package server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"gitlab.com/skjutare/tea-eyes/internal/render"
)

// RenderImageMetadata is the structured metadata returned next to the image
// content block.
type RenderImageMetadata struct {
	Format    string `json:"format"`
	Width     int    `json:"width"`  // requested columns, not pixel width
	Height    int    `json:"height"` // requested rows
	Mime      string `json:"mime"`
	Bytes     int    `json:"bytes"`
	CacheHit  bool   `json:"cache_hit"`
	CachePath string `json:"cache_path"`
}

func registerRenderImage(s *server.MCPServer, r *render.Renderer) {
	tool := mcp.NewTool("tui_render_image",
		mcp.WithDescription(
			"Render a TUI command as a PNG or GIF via Charm's VHS and return "+
				"the image to Claude. Use this when text capture is not enough — "+
				"to judge colors, spacing, borders, or typography.",
		),
		mcp.WithString("command", mcp.Required(),
			mcp.Description("Binary or shell command to run.")),
		mcp.WithArray("args", mcp.Description("Arguments to pass to the command."),
			mcp.Items(map[string]any{"type": "string"})),
		mcp.WithArray("keys", mcp.Description(
			"Key sequence to send after spawn. Same notation as tui_capture_text."),
			mcp.Items(map[string]any{"type": "string"})),
		mcp.WithNumber("width", mcp.Description("Terminal width in columns."), mcp.DefaultNumber(80)),
		mcp.WithNumber("height", mcp.Description("Terminal height in rows."), mcp.DefaultNumber(24)),
		mcp.WithString("font_family", mcp.Description("Monospace font name."),
			mcp.DefaultString("JetBrains Mono")),
		mcp.WithNumber("font_size", mcp.Description("Font size in points."), mcp.DefaultNumber(14)),
		mcp.WithString("theme", mcp.Description("VHS theme name (see `vhs themes`)."),
			mcp.DefaultString("Dracula")),
		mcp.WithString("format", mcp.Description("Output format."),
			mcp.DefaultString("png"), mcp.Enum("png", "gif")),
		mcp.WithNumber("padding", mcp.Description("Pixel padding around the terminal."),
			mcp.DefaultNumber(20)),
		mcp.WithNumber("settle_ms", mcp.Description("Milliseconds to wait between key sends."),
			mcp.DefaultNumber(300)),
		mcp.WithBoolean("no_cache", mcp.Description("Bypass the render cache."),
			mcp.DefaultBool(false)),
		mcp.WithString("cwd", mcp.Description("Working directory for the spawned command.")),
	)
	s.AddTool(tool, makeRenderImageHandler(r))
}

func makeRenderImageHandler(r *render.Renderer) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		command, _ := args["command"].(string)
		if command == "" {
			return mcp.NewToolResultError("tui_render_image: 'command' is required"), nil
		}
		strArgs, err := stringSlice(args, "args")
		if err != nil {
			return mcp.NewToolResultError("tui_render_image: " + err.Error()), nil
		}
		keySpecs, err := stringSlice(args, "keys")
		if err != nil {
			return mcp.NewToolResultError("tui_render_image: " + err.Error()), nil
		}

		opts := render.RenderOpts{
			Command:    command,
			Args:       strArgs,
			Keys:       keySpecs,
			Width:      intArg(args, "width", 80),
			Height:     intArg(args, "height", 24),
			FontFamily: stringArg(args, "font_family", "JetBrains Mono"),
			FontSize:   intArg(args, "font_size", 14),
			Theme:      stringArg(args, "theme", "Dracula"),
			Format:     stringArg(args, "format", "png"),
			Padding:    intArg(args, "padding", 20),
			SettleMs:   intArg(args, "settle_ms", 300),
			Cwd:        stringArg(args, "cwd", ""),
		}
		noCache := boolArg(args, "no_cache", false)

		out, err := r.Render(ctx, opts, noCache)
		if err != nil {
			return mcp.NewToolResultErrorFromErr("tui_render_image: render failed", err), nil
		}

		meta := RenderImageMetadata{
			Format:    out.Format,
			Width:     opts.Width,
			Height:    opts.Height,
			Mime:      out.Mime,
			Bytes:     len(out.Bytes),
			CacheHit:  out.CacheHit,
			CachePath: out.CachePath,
		}
		metaJSON, _ := json.MarshalIndent(meta, "", "  ")
		fallback := fmt.Sprintf("rendered %s (%d bytes, cache_hit=%v)\n%s",
			out.Format, len(out.Bytes), out.CacheHit, string(metaJSON))

		result := mcp.NewToolResultImage(fallback,
			base64.StdEncoding.EncodeToString(out.Bytes), out.Mime)
		result.StructuredContent = meta
		return result, nil
	}
}

func stringArg(args map[string]any, key, def string) string {
	v, ok := args[key]
	if !ok || v == nil {
		return def
	}
	s, ok := v.(string)
	if !ok || s == "" {
		return def
	}
	return s
}
