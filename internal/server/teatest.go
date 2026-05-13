package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"

	"gitlab.com/skjutare/tea-eyes/internal/teatest"
)

// rejectModeArg returns an actionable error if the caller passed mode,
// tmux_session, or tmux_persist to a tool that runs in-process via teatest.
// teatest doesn't use a driver, so those parameters are nonsensical there.
func rejectModeArg(args map[string]any, toolName string) error {
	if v, ok := args["mode"]; ok && v != nil && v != "" {
		return fmt.Errorf(
			"%s: 'mode' is not applicable — this tool runs the model in-process "+
				"via teatest and does not spawn a separate terminal session. Use "+
				"tui_capture_text or tui_render_image if you need pty/tmux mode",
			toolName,
		)
	}
	for _, k := range [2]string{"tmux_session", "tmux_persist"} {
		if v, ok := args[k]; ok && v != nil && v != "" && v != false {
			return fmt.Errorf("%s: %q is not applicable to teatest-based tools", toolName, k)
		}
	}
	return nil
}

// TestGoldenResult is the structured payload returned by tui_test_golden.
type TestGoldenResult struct {
	Match              bool   `json:"match"`
	Created            bool   `json:"created"`
	Diff               string `json:"diff,omitempty"`
	GoldenPath         string `json:"golden_path"`
	FinalOutputPreview string `json:"final_output_preview"`
}

// InspectModelResult is the structured payload returned by tui_inspect_model.
type InspectModelResult struct {
	ModelJSON          string `json:"model_json"`
	ViewText           string `json:"view_text"`
	FinalOutputPreview string `json:"final_output_preview"`
}

func registerTestGolden(s *server.MCPServer, driver *teatest.Driver) {
	tool := mcp.NewTool("tui_test_golden",
		mcp.WithDescription(
			"Run a Bubble Tea program in-process via the teatest harness and "+
				"compare its final output against a golden file. The package must "+
				"export TeaEyesNewModel() tea.Model under the `teaeyes` build tag.",
		),
		mcp.WithString("package_path",
			mcp.Description("Filesystem path to the Bubble Tea package directory."),
			mcp.Required(),
		),
		mcp.WithString("model_func",
			mcp.Description("Exported constructor name. Defaults to TeaEyesNewModel."),
		),
		mcp.WithArray("keys",
			mcp.Description("Key sequence to send before capturing final output."),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithNumber("width", mcp.Description("Terminal width in columns."), mcp.DefaultNumber(80)),
		mcp.WithNumber("height", mcp.Description("Terminal height in rows."), mcp.DefaultNumber(24)),
		mcp.WithString("color_profile",
			mcp.Description("Color profile: Ascii (default, deterministic), ANSI, ANSI256, or TrueColor."),
		),
		mcp.WithString("golden_file",
			mcp.Description("Path to the golden file. Created on first run."),
			mcp.Required(),
		),
		mcp.WithBoolean("update_golden",
			mcp.Description("If true, overwrite the golden with the current output."),
			mcp.DefaultBool(false),
		),
	)
	s.AddTool(tool, makeTestGoldenHandler(driver))
}

func makeTestGoldenHandler(driver *teatest.Driver) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()

		if err := rejectModeArg(args, "tui_test_golden"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		pkg, _ := args["package_path"].(string)
		if pkg == "" {
			return mcp.NewToolResultError("tui_test_golden: 'package_path' is required"), nil
		}
		golden, _ := args["golden_file"].(string)
		if golden == "" {
			return mcp.NewToolResultError("tui_test_golden: 'golden_file' is required"), nil
		}
		modelFunc, _ := args["model_func"].(string)
		colorProfile, _ := args["color_profile"].(string)
		keys, err := stringSlice(args, "keys")
		if err != nil {
			return mcp.NewToolResultError("tui_test_golden: " + err.Error()), nil
		}
		width := intArg(args, "width", 80)
		height := intArg(args, "height", 24)
		update := boolArg(args, "update_golden", false)

		out, err := driver.RunGolden(ctx, teatest.GoldenOpts{
			PackagePath:  pkg,
			ModelFunc:    modelFunc,
			Keys:         keys,
			Width:        width,
			Height:       height,
			ColorProfile: colorProfile,
			GoldenFile:   golden,
			UpdateGolden: update,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("tui_test_golden", err), nil
		}

		summary := "match"
		if out.Created {
			summary = "created golden " + out.GoldenPath
		} else if !out.Match {
			summary = "mismatch:\n" + out.Diff
		}

		return mcp.NewToolResultStructured(TestGoldenResult{
			Match:              out.Match,
			Created:            out.Created,
			Diff:               out.Diff,
			GoldenPath:         out.GoldenPath,
			FinalOutputPreview: preview(out.FinalOutput, 500),
		}, summary), nil
	}
}

func registerInspectModel(s *server.MCPServer, driver *teatest.Driver) {
	tool := mcp.NewTool("tui_inspect_model",
		mcp.WithDescription(
			"Drive a Bubble Tea program in-process via teatest and return the "+
				"JSON-encoded exported fields of its final model plus the current "+
				"View() output. Unexported fields are not visible — add debug "+
				"accessors if you need them.",
		),
		mcp.WithString("package_path",
			mcp.Description("Filesystem path to the Bubble Tea package directory."),
			mcp.Required(),
		),
		mcp.WithString("model_func",
			mcp.Description("Exported constructor name. Defaults to TeaEyesNewModel."),
		),
		mcp.WithArray("keys",
			mcp.Description("Key sequence to send before snapshotting."),
			mcp.Items(map[string]any{"type": "string"}),
		),
		mcp.WithNumber("width", mcp.Description("Terminal width in columns."), mcp.DefaultNumber(80)),
		mcp.WithNumber("height", mcp.Description("Terminal height in rows."), mcp.DefaultNumber(24)),
		mcp.WithString("color_profile",
			mcp.Description("Color profile (default Ascii)."),
		),
	)
	s.AddTool(tool, makeInspectModelHandler(driver))
}

func makeInspectModelHandler(driver *teatest.Driver) server.ToolHandlerFunc {
	return func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		args := req.GetArguments()
		if err := rejectModeArg(args, "tui_inspect_model"); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		pkg, _ := args["package_path"].(string)
		if pkg == "" {
			return mcp.NewToolResultError("tui_inspect_model: 'package_path' is required"), nil
		}
		modelFunc, _ := args["model_func"].(string)
		colorProfile, _ := args["color_profile"].(string)
		keys, err := stringSlice(args, "keys")
		if err != nil {
			return mcp.NewToolResultError("tui_inspect_model: " + err.Error()), nil
		}
		width := intArg(args, "width", 80)
		height := intArg(args, "height", 24)

		out, err := driver.Inspect(ctx, teatest.InspectOpts{
			PackagePath:  pkg,
			ModelFunc:    modelFunc,
			Keys:         keys,
			Width:        width,
			Height:       height,
			ColorProfile: colorProfile,
		})
		if err != nil {
			return mcp.NewToolResultErrorFromErr("tui_inspect_model", err), nil
		}

		return mcp.NewToolResultStructured(InspectModelResult{
			ModelJSON:          out.ModelJSON,
			ViewText:           out.ViewText,
			FinalOutputPreview: preview(out.FinalOutput, 500),
		}, out.ViewText), nil
	}
}

func preview(s string, n int) string {
	s = strings.TrimRight(s, "\n")
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
