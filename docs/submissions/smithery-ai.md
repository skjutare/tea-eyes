# Submission draft — smithery.ai

**Listing type:** MCP server

**Name:** tea-eyes

**Tagline:** Playwright for terminals — visual feedback for TUI development.

**Server command:**

```sh
tea-eyes serve
```

**Transport:** stdio

**Tools exposed:**

| Tool | Inputs (required only) | What it does |
|------|------------------------|---------------|
| `tui_capture_text` | `command` | Spawn TUI under pty/tmux, send keys, return text grid |
| `tui_render_image` | `command` | Render TUI as PNG/GIF via VHS, return as MCP image content block |
| `tui_test_golden` | `package_path`, `golden_file` | Drive a Bubble Tea model in-process via teatest, compare against golden file |
| `tui_inspect_model` | `package_path` | Drive a Bubble Tea model, return exported fields as JSON + `View()` |
| `tui_session_attach_hint` | `session_name` | Return `tmux attach -t …` command for a persistent capture session |

**External dependencies (runtime, optional per-tool):**

- `vhs` + `ttyd` + `ffmpeg` — required for `tui_render_image`
- `tmux` — required for `mode="tmux"` on `tui_capture_text`
- Go toolchain — required for `tui_test_golden` / `tui_inspect_model`
  (compiles a per-package harness)

`tea-eyes doctor` reports each dependency's status with an install hint.

**Install:**

```sh
go install gitlab.com/skjutare/tea-eyes/cmd/tea-eyes@latest
```

Or download a release binary from
https://gitlab.com/skjutare/tea-eyes/-/releases (linux/darwin/windows ×
amd64/arm64).

**Repository:** https://gitlab.com/skjutare/tea-eyes

**License:** MIT

**Author:** Christoffer Skjutare
