# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Status

**Phases 0–7 complete.** Repo scaffold, license, CI; MCP server with `tui_capture_text` (pty + tmux), `tui_render_image` (VHS), `tui_test_golden` and `tui_inspect_model` (teatest white-box), `tui_session_attach_hint`; `tea-eyes doctor`/`cache`/`serve` subcommands; both skills (`tea-eyes-loop`, `tea-eyes-bubbletea`); both subagents (`tui-designer`, `tui-tester`) and plugin manifest; release engineering (`.goreleaser.yaml`, GitHub + GitLab release workflows, ldflags-stamped version); full docs (`architecture.md`, `workflow.md` narrative tutorial, `mcp-tools.md` reference, `agents.md`, `roadmap.md`); submission drafts for the five target directories; launch announcement draft.

**What remains before v0.1.0 ships:** (a) cut the actual `v0.1.0` tag, move `[Unreleased]` → `[0.1.0] - YYYY-MM-DD` in CHANGELOG, push to GitLab so CI publishes artifacts and the mirror picks it up; (b) flip both repos public; (c) re-record the demo PNG/GIF from a real Claude session via `make demo-render`; (d) work through the manual QA pass in `docs/build-prompts/phase-7-release.md` §8; (e) actually submit to the five directories using the drafts in `docs/submissions/`.

The authoritative plan is `docs/plan.md`. Per-phase prompts are `docs/build-prompts/phase-{0..7}-*.md`. When the plan and a per-phase prompt disagree, **the per-phase prompt wins** (it's more specific); update `docs/plan.md` to match.

## Project: tea-eyes

A Claude Code plugin (Go MCP server + skills + subagents) that gives Claude visual feedback when developing TUIs. "Playwright for TUI development." First-class support for the Charm ecosystem (Bubble Tea, Lipgloss, VHS, teatest).

- **Author:** Christoffer Skjutare
- **License:** MIT (matches Charm)
- **Canonical repo:** `gitlab.com/skjutare/tea-eyes` (mirrored to `github.com/skjutare/tea-eyes`, both private during 0.x)
- **Go version:** 1.26 minimum
- **MCP framework:** `github.com/mark3labs/mcp-go`

## Architecture (target end state)

```
Claude Code ──stdio MCP──► tea-eyes server (Go)
                              ├── tui_capture_text   ─► driver (pty | tmux) ─► user TUI
                              ├── tui_render_image   ─► vhs subprocess       ─► PNG/GIF
                              ├── tui_test_golden    ─► teatest harness      ─► user pkg
                              ├── tui_inspect_model  ─► teatest harness      ─► user pkg
                              └── tui_session_attach_hint
```

Three complementary capture/render strategies — Claude picks the right one per task, guided by the `tea-eyes-loop` skill:

| Strategy | Tool | Speed | Fidelity | Coupling |
|---|---|---|---|---|
| Text via pty/tmux + VT emulator | `tui_capture_text` | ~50ms | ASCII grid | none |
| Image via VHS | `tui_render_image` | 2–5s | true pixels | needs vhs/ttyd/ffmpeg |
| In-process via teatest | `tui_test_golden`, `tui_inspect_model` | ~10ms | structured | Bubble Tea + build tag |

### White-box pattern (Phase 3)

For teatest tools, the user adds a `TeaEyesNewModel() tea.Model` function under the `teaeyes` build tag. tea-eyes generates a tiny harness binary that imports the user's package and drives the model via teatest. Behavior tests force `Ascii` color profile for stable goldens; color tests use `TrueColor` and assert on SGR markers, not RGB. **Non-negotiable.**

### Target layout (post-bootstrap)

```
cmd/tea-eyes/main.go            # MCP server entry
internal/{server,driver,pty,tmux,capture,render,teatest,keys}/
pkg/teaeyes/                    # public Go API
plugin/{plugin.json,skills/,agents/}
examples/{hello-tui,multi-pane}/
docs/{plan.md,architecture.md,mcp-tools.md,workflow.md,...}
test/{integration,golden}/
```

## Working with phases

Each phase ends in a working, demoable state. Acceptance criteria in each phase prompt are non-negotiable — don't move on until they all pass. Every phase prompt lists explicit **anti-scope** items (deferred features); if tempted to over-build, re-read the anti-scope.

| # | File | Goal |
|---|---|---|
| 0 | `phase-0-bootstrap.md` | Repo skeleton, license, CI |
| 1 | `phase-1-mcp-pty.md` | MCP server + pty + `tui_capture_text` |
| 2 | `phase-2-vhs-render.md` | VHS + `tui_render_image` + caching + `doctor` |
| 3 | `phase-3-teatest.md` | `tui_test_golden`, `tui_inspect_model`, white-box pattern |
| 4 | `phase-4-skills.md` | Two SKILL.md + plugin manifest |
| 5 | `phase-5-subagents.md` | `tui-designer`, `tui-tester` |
| 6 | `phase-6-tmux.md` | tmux driver + `mode` parameter |
| 7 | `phase-7-release.md` | Polish, narrative tutorial, v0.1.0 |

Commands like `make build test lint` are introduced in Phase 0 — consult that phase prompt before assuming any toolchain command exists.

## Composition (not replacement)

tea-eyes is **additive** to the existing ecosystem:

- **GGPrompts/TFE bubbletea skill** owns layout rules (the 4 Golden Rules). `tea-eyes-bubbletea` defers to it — do not duplicate layout guidance.
- **jmlago "Debug TUIs with tmux" skill** — coexists; tea-eyes formalizes the same pattern as a typed MCP tool surface.
- **Charm (Bubble Tea, Lipgloss, VHS, teatest)** — the foundation. tea-eyes is a thin orchestration layer; credit explicitly in NOTICE/README.

## Cross-cutting rules

- **Error messages must be actionable.** Bad: "exit 1". Good: name the missing binary, the PATH searched, and a fix command.
- **Caching:** VHS renders and teatest harness binaries cached under `$XDG_CACHE_HOME/tea-eyes/` keyed by SHA-256 of canonicalized inputs. Last-write-wins; no locks.
- **No telemetry. Ever.** Listed as a permanent non-goal in `plan.md` §2.
- **No new TUI framework support in v1.** Ratatui/Textual/Ink are v2 candidates.
- **SemVer from v0.1.0.** Breaking changes allowed in 0.x minors but must be called out in CHANGELOG.
