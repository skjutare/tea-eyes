# Changelog

All notable changes to tea-eyes will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added

- `internal/tmux` package — tmux CLI wrapper supporting create/attach/kill
  session, raw-byte key delivery via `send-keys -H`, `capture-pane -e`,
  resize, and `respawn-pane` for reusing a persistent session across
  successive calls (Phase 6).
- `internal/driver` (`capturedrv`) package — backend-agnostic Driver
  interface and `Mode` enum (`pty`, `tmux`) used by MCP tools. Pty stays
  the default for backward compatibility (Phase 6).
- `tui_capture_text` now accepts `mode`, `tmux_session`, and `tmux_persist`
  parameters. The structured result echoes the chosen `mode` and, when
  applicable, the `tmux_session` name so a follow-up call (or the user)
  can attach (Phase 6).
- `tui_session_attach_hint` MCP tool — returns the `tmux attach -t …`
  command for a given session name, plus an `exists` flag (Phase 6).
- `tui_render_image` accepts `mode` for parameter symmetry but rejects
  `mode="tmux"` with an actionable error explaining the VHS/tmux gap.
  `tui_test_golden` and `tui_inspect_model` reject `mode` entirely since
  they run in-process via teatest (Phase 6).
- `tea-eyes doctor` now reports tmux availability as an optional
  dependency, distinguishing missing-required from missing-optional.
- `docs/mcp-tools.md` gains a "Choosing a mode" section, the new
  parameters on `tui_capture_text`, and a `tui_session_attach_hint`
  reference. `docs/workflow.md` gains a "Watching Claude drive the TUI"
  section with a worked example.
- `plugin/agents/tui-designer.md` — reference subagent for iterating on
  the visual design of TUIs. Renders the TUI as an image, describes
  what it sees, proposes one focused change at a time, and verifies via
  re-render (Phase 5).
- `plugin/agents/tui-tester.md` — reference subagent for golden-file
  testing of Bubble Tea apps. Enforces the `TeaEyesNewModel` white-box
  pattern and the ASCII-vs-TrueColor color-profile discipline for
  stable, reviewable goldens (Phase 5).
- `docs/agents.md` — when to invoke each subagent, example prompts that
  trigger them, the tool lists (including tools they deliberately
  *don't* have), and how they compose with the skills.
- Plugin manifest (`plugin/plugin.json`) now registers both subagents
  under `agents` (the entries were already present from Phase 4; the
  files they reference now exist).
- `plugin/skills/tea-eyes-loop/SKILL.md` — framework-agnostic skill that
  teaches the capture → reason → edit → re-capture loop and when to pick
  which tea-eyes MCP tool (Phase 4).
- `plugin/skills/tea-eyes-bubbletea/SKILL.md` — Bubble Tea-specific skill
  covering the `TeaEyesNewModel` white-box pattern, color-profile
  discipline for stable goldens, and the teatest workflow. Strictly
  additive to the GGPrompts/TFE bubbletea skill (Phase 4).
- `plugin/plugin.json` — Claude Code plugin manifest wiring the
  stdio-served `tea-eyes` MCP server, both skills, and (forward-referenced)
  subagent paths.
- `tea-eyes serve` subcommand — explicit entrypoint for the stdio MCP
  server; the bare `tea-eyes` invocation still serves for backward compat.
- `docs/compat-ggprompts.md` — explains how tea-eyes composes with the
  GGPrompts/TFE bubbletea skill (verifier vs. design rules).
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
