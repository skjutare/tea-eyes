# Changelog

All notable changes to tea-eyes will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `tui_render_image` MCP tool — render a TUI as PNG or GIF via Charm's VHS
  and return it as an MCP image content block (Phase 2).
- `internal/render` package — tape-file generation, VHS subprocess
  invocation, and a SHA-256-keyed on-disk cache under
  `$XDG_CACHE_HOME/tea-eyes/renders/`.
- `tea-eyes doctor` subcommand — verifies that `vhs`, `ttyd`, and `ffmpeg`
  are present on `PATH` and reports their versions (or install hints).
- `tea-eyes cache clean` / `tea-eyes cache path` subcommands for managing
  the render cache.
- `examples/multi-pane` — a Bubble Tea demo with two focusable panes and a
  status bar, used as the visual-rendering smoke test.
- `make test-integration` and `make demo-render` targets; the new render
  integration test is guarded by the `integration` build tag.
- `tui_capture_text` MCP tool — spawn a TUI under a pty, optionally drive it
  with keystrokes, and return the rendered text grid (Phase 1).
- `internal/pty` driver with auto-reply for common terminal capability
  queries (CSI DSR, OSC 10/11, primary/secondary DA) so termenv-based TUIs
  render on startup.
- `internal/keys` key-spec parser supporting literals, special keys
  (`enter`, `tab`, arrows, function keys, etc.) and modifiers (`ctrl+`,
  `alt+`, `shift+`, `meta+`).
- `internal/capture` virtual terminal renderer (hinshun/vt10x).
- `examples/hello-tui` Bubble Tea demo used by the integration test.
- End-to-end integration test (`test/integration`) driving the MCP server
  in-process against the example.
