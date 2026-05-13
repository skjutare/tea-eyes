# tea-eyes — Master Build Plan

**Author:** Christoffer Skjutare
**License:** MIT (matching the Charm ecosystem)
**Status:** Phases 0–2 complete; Phase 3 (teatest) up next
**Last updated:** 2026-05-12

This is the single-source-of-truth plan for building **tea-eyes**, a Claude
Code plugin that gives Claude visual feedback when developing terminal user
interfaces. Every per-phase prompt in `docs/build-prompts/` references this
document; keep it in the repo as `docs/plan.md` so Claude can re-read it
whenever it needs orientation.

---

## 1. Vision

**Playwright is to frontend development what tea-eyes is to TUI development.**

When Claude Code builds a web app, it can use Playwright (or the Playwright
MCP) to render the page, see the DOM, click buttons, screenshot the result,
and iterate. When Claude Code builds a TUI, it has historically been blind:
it writes Bubble Tea code, hopes the layout works, and the user has to run
the app and report back.

tea-eyes closes that loop. It exposes an MCP server with three complementary
capture/render strategies, two reference subagents, and two skills — so
Claude can:

- **See** what the TUI renders, as ASCII *or* as actual pixels
- **Drive** the TUI by sending keystrokes
- **Lock in** behaviors with golden-file tests
- **Inspect** in-process state for Bubble Tea apps specifically

The goal is parity-of-experience with Playwright for the iterative design
loop: render → look → reason → edit → re-render.

---

## 2. Scope and non-goals

### In scope for v1 (v0.1.0)

- Framework-agnostic capture/render of any TUI binary
- Bubble Tea-specific in-process driver via teatest
- VHS-based PNG/GIF rendering with deterministic caching
- Both pty (default) and tmux process drivers
- Two reference subagents (`tui-designer`, `tui-tester`)
- Two skills (one framework-agnostic, one Bubble Tea-specific)
- Distribution as a Claude Code plugin
- Cross-platform release (linux/darwin/windows × amd64/arm64)
- Composition with the existing GGPrompts/TFE bubbletea skill

### Explicitly out of scope for v1

- Ratatui (Rust), Textual (Python), Ink (JS) framework plugins — v2 candidates
- Cross-frame visual diffing
- Mutation or property-based testing
- Cross-host operation (SSH, remote tmux)
- Multi-pane / multi-window tmux orchestration
- Recording and replaying user sessions
- A web dashboard / GUI
- Telemetry or analytics of any kind

### Non-goals (will not do, ever)

- Replace or duplicate the GGPrompts/TFE bubbletea skill's layout rules
- Become a TUI framework itself
- Lock users into a specific TUI library

---

## 3. Architecture

### Component diagram

```
                        ┌─────────────────────┐
                        │    Claude Code      │
                        │ (or any MCP client) │
                        └──────────┬──────────┘
                                   │ stdio (JSON-RPC)
                                   ▼
                        ┌─────────────────────┐
                        │  tea-eyes server    │
                        │  (Go, mark3labs/    │
                        │   mcp-go)           │
                        └──────────┬──────────┘
                                   │
       ┌───────────────────────────┼───────────────────────────┐
       │                           │                           │
       ▼                           ▼                           ▼
┌─────────────┐            ┌──────────────┐            ┌──────────────┐
│  driver/    │            │  render/     │            │  teatest/    │
│  pty | tmux │            │  vhs wrapper │            │  harness gen │
└──────┬──────┘            └──────┬───────┘            └──────┬───────┘
       │                          │                           │
       ▼                          ▼                           ▼
┌─────────────┐            ┌──────────────┐            ┌──────────────┐
│  user TUI   │            │  vhs binary  │            │  user pkg    │
│  (any lang) │            │  (subprocess)│            │  built with  │
└─────────────┘            └──────────────┘            │  -tags       │
                                                       │  teaeyes     │
                                                       └──────────────┘
```

### Three capture/render strategies

| Strategy | Tool | Speed | Fidelity | Coupling |
|----------|------|-------|----------|----------|
| Text capture via pty/tmux + VT emulator | `tui_capture_text` | fast (~50ms) | ASCII grid only | none |
| Image render via VHS | `tui_render_image` | slow (~2-5s) | true pixels | requires vhs/ttyd/ffmpeg |
| In-process via teatest | `tui_test_golden`, `tui_inspect_model` | fastest (~10ms) | full structured output | Bubble Tea + build tag |

Claude picks the right tool for the job, guided by the `tea-eyes-loop` skill.

### MCP tool surface (final, after Phase 6)

| Tool | Purpose | Available after |
|------|---------|-----------------|
| `tui_capture_text` | Capture rendered text from any TUI | Phase 1 |
| `tui_render_image` | Render PNG/GIF of any TUI via VHS | Phase 2 |
| `tui_test_golden` | Bubble Tea golden-file test | Phase 3 |
| `tui_inspect_model` | Bubble Tea model state dump | Phase 3 |
| `tui_session_attach_hint` | Get tmux attach command | Phase 6 |

All process-driving tools accept a `mode: "pty" | "tmux"` parameter (Phase 6),
defaulting to `pty`. teatest tools ignore mode (always in-process).

### Process model

- **pty (default)**: spawn the user's TUI binary in a pseudo-terminal owned by
  the MCP server. Clean, portable, no dependencies. Ephemeral per call.
- **tmux (opt-in)**: drive a tmux session. Optionally persistent across calls.
  Lets the user attach in another terminal and watch Claude work.
- **in-process (Bubble Tea only)**: compile a generated harness with the user's
  package; drive the model via teatest. Fastest, deepest introspection,
  framework-coupled.

### Repository layout (target end state)

```
tea-eyes/
├── LICENSE                           # MIT
├── NOTICE                            # credits to prior art
├── README.md
├── CONTRIBUTING.md
├── CHANGELOG.md
├── Makefile
├── go.mod
├── .goreleaser.yaml                  # Phase 7
├── .gitlab-ci.yml                    # primary CI
├── .github/workflows/                # mirror CI for GitHub
│
├── cmd/
│   └── tea-eyes/
│       └── main.go                   # MCP server entry point
│
├── internal/
│   ├── server/                       # MCP server wiring
│   ├── driver/                       # driver interface (Phase 6)
│   ├── pty/                          # pty-based driver (Phase 1)
│   ├── tmux/                         # tmux-based driver (Phase 6)
│   ├── capture/                      # text capture / VT emulation
│   ├── render/                       # vhs wrapping for PNG/GIF (Phase 2)
│   ├── teatest/                      # bubble tea harness (Phase 3)
│   └── keys/                         # key string parser
│
├── pkg/
│   └── teaeyes/                      # public Go API for embedders
│
├── plugin/                           # Claude Code plugin manifest
│   ├── plugin.json
│   ├── skills/
│   │   ├── tea-eyes-loop/SKILL.md         # framework-agnostic
│   │   └── tea-eyes-bubbletea/SKILL.md    # additive to GGPrompts
│   └── agents/
│       ├── tui-designer.md
│       └── tui-tester.md
│
├── examples/
│   ├── hello-tui/                    # minimal Bubble Tea app
│   └── multi-pane/                   # exercises layout primitives
│
├── docs/
│   ├── plan.md                       # THIS FILE
│   ├── architecture.md
│   ├── mcp-tools.md
│   ├── workflow.md
│   ├── white-box-pattern.md          # the TeaEyesNewModel convention
│   ├── compat-ggprompts.md
│   ├── agents.md
│   ├── roadmap.md
│   ├── img/                          # screenshots, GIFs
│   ├── submissions/                  # directory submission drafts
│   └── build-prompts/                # the per-phase prompts
│
└── test/
    ├── integration/
    └── golden/
```

---

## 4. Phase overview

| # | Name | Goal | Effort | Exit state |
|---|------|------|--------|-----------|
| 0 | Bootstrap ✅ | Repo skeleton, license, CI, docs placeholders | ½ day | Empty but well-formed repo, CI green |
| 1 | MCP + pty ✅ | Server skeleton, key parser, pty driver, `tui_capture_text` | 1–2 days | Claude can capture hello-tui as text |
| 2 | VHS render ✅ | `tui_render_image`, caching, `tea-eyes doctor` | 1–2 days | Claude can see the TUI as PNG |
| 3 | teatest | `tui_test_golden`, `tui_inspect_model`, white-box pattern | 1–2 days | Claude can write & run golden tests |
| 4 | Skills | Two SKILL.md files, plugin manifest, GGPrompts compat doc | ½–1 day | Claude auto-triggers tea-eyes appropriately |
| 5 | Subagents | `tui-designer`, `tui-tester`, agents documentation | ½ day | Named subagents work end-to-end |
| 6 | tmux | tmux driver, `mode` parameter on tools | 1 day | Optional tmux mode with persistent sessions |
| 7 | Release | Polish docs, narrative tutorial, v0.1.0 release | 1–2 days | Tagged release, distribution drafts ready |

**Total: ~7–10 working days.**

Each phase ends in a demoable state. Stop after any phase if priorities
change; everything built up to that point is independently useful.

---

## 5. Per-phase summaries

Detailed prompts are in `docs/build-prompts/phase-N-*.md`. The summaries
below exist so this single document orients anyone (human or agent) to the
full build.

### Phase 0 — Bootstrap

Empty Go module on `gitlab.com/<namespace>/tea-eyes` (mirrored to GitHub).
MIT license matching Charm. README states what tea-eyes is, what it isn't,
and credits prior art. CI pipeline runs `vet`, `staticcheck`, `test`,
goreleaser snapshot. No application code.

**Exit:** `make build test lint` green; CI green; tree matches the layout in
section 3.

### Phase 1 — MCP + pty

`mark3labs/mcp-go` server registered with one tool: `tui_capture_text`.
Wraps `creack/pty` and a VT100 emulator (`charmbracelet/x/ansi` or
`hinshun/vt10x`). Includes a key parser that handles plain runes, special
keys, and modifiers (`ctrl+c`, `alt+x`, etc.). Ships `examples/hello-tui`
as the smoke-test target.

**Exit:** From a fresh Claude Code session with tea-eyes connected, asking
"show me what `examples/hello-tui` renders" returns the actual rendered
text.

### Phase 2 — VHS render

Adds `tui_render_image` tool. Generates `.tape` files from interaction
specs, shells out to `vhs`, returns PNG/GIF as MCP image content blocks
so Claude sees them natively. Caches rendered files keyed by hash of
inputs. Adds `tea-eyes doctor` subcommand to verify external dependencies
(vhs, ttyd, ffmpeg). Adds `examples/multi-pane` as a layout demo.

**Exit:** Claude can render `examples/multi-pane` as a PNG, describe the
layout in natural language, suggest a change, edit code, re-render, and
visually confirm the change.

### Phase 3 — teatest

Adds `tui_test_golden` and `tui_inspect_model` tools. Uses the
**white-box pattern**: user adds a `TeaEyesNewModel() tea.Model` function
under the `teaeyes` build tag; tea-eyes generates a tiny harness binary
that imports the user's package and drives the model via teatest. Forces
ASCII color profile by default for stable goldens.

**Exit:** Claude can write a golden test for `examples/hello-tui` and the
test stays stable across runs.

### Phase 4 — Skills

Two SKILL.md files:
- `tea-eyes-loop` — framework-agnostic, teaches the visual feedback loop and
  tool selection.
- `tea-eyes-bubbletea` — Bubble Tea-specific, **strictly additive** to the
  GGPrompts/TFE bubbletea skill. Defers all layout rules to GGPrompts;
  adds only tea-eyes-specific workflow guidance (white-box pattern, color
  profile, image rendering for design review).

Plus `plugin/plugin.json` declaring the MCP server, skills, and (forward-
referenced) subagents. Plus `docs/compat-ggprompts.md`.

**Exit:** Mentioning a TUI design task in Claude Code automatically
triggers the right skill.

### Phase 5 — Subagents

Two reference subagents:
- `tui-designer` — iterates on visual design via the render-look-edit loop.
  Hard rule: never declare done without a final image render.
- `tui-tester` — writes and maintains golden tests. Enforces white-box
  pattern, color profile discipline, naming conventions.

**Exit:** Natural-language prompts like "tweak the focus color of the right
panel" automatically invoke `tui-designer`.

### Phase 6 — tmux

Adds tmux as an alternative process driver. All process-driving tools
gain `mode`, `tmux_session`, `tmux_persist` parameters. Adds
`tui_session_attach_hint` tool so Claude can tell the user how to watch.
`tui_render_image` rejects `mode=tmux` with a clear error (rendering from
tmux is out of scope).

**Exit:** Claude can run a TUI in a persistent tmux session that the user
attaches to in another terminal and watches live.

### Phase 7 — Release

Long-form narrative tutorial in `docs/workflow.md` with embedded vhs-
generated screenshots. Architecture doc with Mermaid diagram. Final pass on
README. goreleaser configured for cross-platform release. Tag v0.1.0.
Submission drafts for mcpmarket, claude-plugins.dev, fastmcp.me,
smithery.ai. Launch announcement draft mirroring Hatchet's framing.

**Exit:** v0.1.0 tagged on GitLab and GitHub with artifacts; submissions
ready to send; manual QA pass complete.

---

## 6. Composition with the existing ecosystem

tea-eyes is **additive**. It does not replace any of the following — it
extends them.

### GGPrompts/TFE bubbletea skill

Encodes layout rules (the "4 Golden Rules": account for borders, never
auto-wrap in bordered panels, match mouse detection to layout, use weights
not pixels). tea-eyes-bubbletea **defers** all layout rules to this skill.
The README and `docs/compat-ggprompts.md` recommend installing both.

### jmlago "Debug TUIs with tmux" skill

The original tmux-driven TUI inspection workflow. tea-eyes formalizes the
same pattern as a first-class MCP server with a typed tool surface. The
two coexist — users can use jmlago's skill for ad-hoc tmux work and
tea-eyes when they want the structured workflow.

### rigerc/bubbletea-v2-scaffold

A Bubble Tea v2 project scaffold. tea-eyes is plugin-shaped; the scaffold
is project-shaped. They compose — start from the scaffold, install
tea-eyes as a Claude Code plugin.

### Charm ecosystem (Bubble Tea, Lipgloss, Bubbles, VHS, teatest)

The foundation. tea-eyes is a thin orchestration layer; all the heavy
lifting belongs to Charm. License matches (MIT), credits are explicit,
and the project naming pays homage (`tea-eyes` from "reading tea leaves").

### Hatchet's "Building a TUI is easy now" blog post

Validated the approach (Claude + tmux + Bubble Tea is genuinely productive).
The launch announcement explicitly credits and links it.

---

## 7. Cross-cutting decisions

These apply across all phases. Document them once here so per-phase prompts
don't have to re-derive them.

### MCP framework

`github.com/mark3labs/mcp-go`. Most popular Go MCP framework, simple to
use, well-documented, wide contributor pool. Anthropic's official Go SDK
exists but is newer; revisit at v1.0.

### Go version

1.26 minimum. Use the toolchain directive.

### Versioning

SemVer from v0.1.0. v1.0.0 only after the MCP tool surface has been stable
for ≥3 months and used by ≥1 external project. During 0.x, breaking
changes are allowed in minor versions but always called out in CHANGELOG.

### Distribution

- **Canonical repo:** `gitlab.com/<namespace>/tea-eyes`
- **Mirror:** `github.com/<namespace>/tea-eyes` (push mirror configured
  in GitLab; reduces discoverability friction since most Claude Code
  skill directories index GitHub)
- **Release artifacts:** built once via goreleaser, published to both
  GitLab Releases and GitHub Releases
- **Module path:** the GitLab one is canonical; document this in README

### License

MIT. Copyright line: `Copyright (c) 2026 Christoffer Skjutare`. NOTICE
file credits Charm (MIT), GGPrompts/TFE skill, jmlago skill, rigerc
scaffold, Hatchet post, mark3labs/mcp-go.

### Issue triage

Side-project SLA: best-effort, no guarantees. Stated explicitly in
CONTRIBUTING.md.

### Color profile defaults

Behavior tests force `Ascii` color profile (decouples from theme changes).
Color-specific tests use `TrueColor` and assert on SGR markers, not RGB
values. This rule is non-negotiable and lives in the
`tea-eyes-bubbletea` skill.

### Caching

VHS renders cached to `$XDG_CACHE_HOME/tea-eyes/renders/` keyed by SHA-256
of canonicalized inputs. teatest harness binaries cached similarly. Both
have CLI commands to clear (`tea-eyes cache clean`).

### Error handling philosophy

Every MCP tool error must be **actionable**. Bad: "exit 1". Good:
"command not found: foo (PATH: /usr/bin:...). Did you build the binary
first? Try `go build ./examples/hello-tui`."

### Concurrency

Each MCP call is independent. No shared state across calls except: the
render cache, the harness binary cache, and (when persistent) tmux
sessions. No locks needed for the caches because last-write-wins on
identical inputs is fine.

---

## 8. Naming history

Working title: **tea-eyes** ("eyes for your tea app"; nods to "reading tea
leaves" / tasseography). Alternates considered: `tasseo`, `scry`,
`leafread`. Locked in at Phase 0; renaming after that requires
search-and-replace plus repo move.

---

## 9. Acceptance criteria for v0.1.0

The release ships when **all** of the following are true:

1. All seven phases complete; their per-phase acceptance criteria all pass.
2. CI green on both GitLab and GitHub.
3. `docs/workflow.md` is a complete narrative tutorial that a stranger can
   follow start-to-finish without external help.
4. The README has a working demo GIF at the top.
5. Every MCP tool returns a sensible response and a useful error in every
   tested failure mode.
6. Manual QA pass complete: every command in `docs/workflow.md` verified
   in a fresh checkout.
7. Cross-platform builds succeed (linux/darwin/windows × amd64/arm64).
8. Submission drafts exist for all five target directories.
9. Launch announcement draft exists.
10. The plugin installs cleanly via `claude plugin install` (or the current
    Claude Code plugin install command).

---

## 10. Roadmap beyond v0.1.0

Lives in `docs/roadmap.md`. Indicative items, not commitments:

- **v0.2** — Ratatui plugin, Textual plugin (framework-aware drivers
  matching the teatest model)
- **v0.3** — visual diff between renders (PNG diff with bounding boxes)
- **v0.4** — record-and-replay user sessions for regression suites
- **v1.0** — API stability commitment after 3+ months of v0.x in use

---

## 11. Working with this plan

If you (a future Claude Code session, or Christoffer himself) come back to
this and feel disoriented, here's how to re-orient:

1. Read sections 1–4 of this document. That's the whole plan in five
   minutes.
2. Check the phase status in the README (which phases are checked off).
3. Open the next phase prompt in `docs/build-prompts/phase-N-*.md`.
4. Each phase prompt is self-contained and lists its own acceptance
   criteria. Don't move on until they all pass.
5. If a phase prompt and this master plan disagree, the per-phase prompt
   wins (it's more specific). Update this plan to match.

When in doubt: the goal is the loop. Render, look, reason, edit, re-render.
Everything else is in service of that loop.
