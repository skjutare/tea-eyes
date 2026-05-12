# Changelog

All notable changes to tea-eyes will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

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
