# Changelog

All notable changes to tea-eyes will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `tui_test_golden` MCP tool — drive a Bubble Tea model in-process via
  teatest and compare its final output against a golden file (Phase 3).
- `tui_inspect_model` MCP tool — drive a Bubble Tea model and return its
  exported fields as JSON plus the current `View()` text (Phase 3).
- `internal/teatest` package — generates a build-tagged `_test.go` harness
  in the user's package, compiles it via `go test -c -tags teaeyes`, and
  caches the resulting binary under `$XDG_CACHE_HOME/tea-eyes/teatest/`.
- `examples/hello-tui/teaeyes.go` — exemplar `TeaEyesNewModel()` white-box
  hook; `examples/hello-tui/teaeyes_test.go` — example of a hand-written
  teatest golden test, runnable via `go test -tags teaeyes`.
- `docs/white-box-pattern.md` — full explanation of the
  `TeaEyesNewModel` convention.
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
