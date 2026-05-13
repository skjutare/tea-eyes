# Architecture

tea-eyes is a Go-based [MCP](https://modelcontextprotocol.io) server that
gives Claude Code (or any MCP client) three complementary ways to *see* a
terminal user interface: text capture, image render, and in-process
introspection of Bubble Tea models.

This document covers the runtime topology, the per-package responsibilities,
the design contracts that hold the system together, and the extension
points designed in for v2 and beyond.

## Runtime topology

```mermaid
flowchart LR
    CC[Claude Code<br/>or any MCP client] -->|stdio JSON-RPC| S[tea-eyes server<br/>cmd/tea-eyes]

    S --> D[driver<br/>pty | tmux]
    S --> R[render<br/>vhs wrapper]
    S --> T[teatest harness<br/>generated _test.go]

    D --> U1[user TUI binary<br/>any language]
    R --> V[vhs subprocess<br/>+ ttyd + ffmpeg]
    V --> U2[user TUI binary]
    T --> U3[user Bubble Tea pkg<br/>build tag: teaeyes]

    R -.cache.-> C1[(XDG_CACHE_HOME/<br/>tea-eyes/renders)]
    T -.cache.-> C2[(XDG_CACHE_HOME/<br/>tea-eyes/teatest)]
```

Each MCP tool call is independent: a fresh pty (or tmux pane), a fresh VHS
process, or a fresh teatest run. The two caches and (when persistent) tmux
sessions are the only shared state across calls. No locks, no leader; the
caches are content-addressed so last-write-wins on identical inputs is fine.

## Per-package responsibilities

| Package | Responsibility | Key types / funcs |
|---------|----------------|--------------------|
| `cmd/tea-eyes` | Process entry point. Subcommand dispatch: `serve`, `doctor`, `cache`. | `main`, `runServer`, `runDoctor`, `runCache` |
| `internal/server` | MCP server wiring. Registers tools, owns the `Version` symbol used by ldflags. | `New`, `Version` |
| `internal/driver` | Backend-agnostic Driver interface and `Mode` enum (`pty`, `tmux`). | `Driver`, `Mode`, `Capture` |
| `internal/pty` | Ephemeral pseudo-terminal capture. Auto-replies to terminal capability queries (DSR, OSC 10/11, primary/secondary DA) so termenv-based TUIs don't block on startup. | `Run`, `replyOnQuery` |
| `internal/tmux` | tmux CLI wrapper for create/attach/kill, raw-byte key delivery via `send-keys -H`, `capture-pane -e`, resize, `respawn-pane` for session reuse. | `NewSession`, `SendKeys`, `Capture`, `Respawn` |
| `internal/capture` | Virtual terminal renderer wrapping hinshun/vt10x to turn raw pty bytes into a clean text grid. | `Render`, `StripANSI` |
| `internal/keys` | Key-spec parser: literals, special keys (`enter`, `tab`, arrows, F1–F12), modifiers (`ctrl+`, `alt+`, `shift+`, `meta+`). | `Parse`, `KeySpec` |
| `internal/render` | VHS tape-file generation, subprocess invocation, SHA-256-keyed on-disk render cache. | `Render`, `DefaultCacheDir`, `Cache` |
| `internal/teatest` | Generates a build-tagged `_test.go` harness in the user's package, compiles via `go test -c -tags teaeyes`, caches the binary. | `Build`, `RunGolden`, `RunInspect` |
| `pkg/teaeyes` | Public Go API for embedders. Intentionally narrow — most users go through the MCP server, not this package. | re-exports of stable surfaces |

## Three capture/render strategies

| Strategy | MCP tool(s) | Typical latency | Fidelity | Coupling |
|---|---|---|---|---|
| Text via pty/tmux + VT emulator | `tui_capture_text` | ~50 ms | ASCII grid | none |
| Image via VHS | `tui_render_image` | 2–5 s cold, ~5 ms cached | true pixels | needs vhs + ttyd + ffmpeg |
| In-process via teatest | `tui_test_golden`, `tui_inspect_model` | ~10 ms (cached binary) | structured grid + JSON state | Bubble Tea + build tag |

The skill `tea-eyes-loop` teaches Claude when to pick which. The two
subagents (`tui-designer`, `tui-tester`) bake the choice into their tool
list — designer gets capture+render, tester gets teatest+capture.

## Design contracts (non-negotiable)

These rules hold the system together across phases. Don't break them
casually.

### 1. Errors are actionable

Every MCP tool error names the failing component, the input that caused it,
and a suggested fix. Bad: `exit 1`. Good:

> `command not found: vhs (PATH: /usr/local/bin:/usr/bin). Install with
> 'brew install vhs' or 'go install github.com/charmbracelet/vhs@latest';
> then re-run 'tea-eyes doctor' to verify.`

This is enforced by code review, not by a linter.

### 2. White-box pattern for teatest

teatest tools never spawn a user binary. They generate a tiny `_test.go`
harness under the `teaeyes` build tag that calls a user-supplied
`TeaEyesNewModel() tea.Model` constructor. See
[`white-box-pattern.md`](./white-box-pattern.md) for the full rationale.

This decision is intentional and load-bearing: it lets tea-eyes drive the
*model*, not the *binary*, which is the only way to get deterministic
in-process behavior.

### 3. Color-profile discipline

Behavior tests force `Ascii` color profile (decouples from theme changes).
Color-specific tests use `TrueColor` and assert on SGR markers, not raw
RGB triplets. Never mix concerns in one golden. The rule lives in the
`tea-eyes-bubbletea` skill and is enforced by the `tui-tester` subagent.

### 4. Caching is content-addressed and lock-free

Both caches (`renders/`, `teatest/`) are keyed by SHA-256 of canonicalized
inputs. Last-write-wins on identical keys; no fsync, no locks, no TTL.
Clear with `tea-eyes cache clean` or by deleting the cache dir.

### 5. No telemetry, ever

Listed in `plan.md` §2 as a permanent non-goal. No analytics, no
phone-home, no opt-in counters. The MCP server only opens stdio and the
subprocesses it spawns.

## Concurrency model

- Each MCP call is independent. The server holds no per-conversation state.
- The render and teatest caches use the filesystem as the synchronization
  primitive — collisions on identical inputs are safe (last-write-wins on
  byte-identical content).
- Persistent tmux sessions are the one piece of cross-call state.
  Identified by session name; tea-eyes never lists or enumerates other
  sessions on the host.

## Extension points (designed for v2+)

The package boundaries above were drawn with these extensions in mind:

### Adding a new process driver

Implement the `internal/driver.Driver` interface (or its v2 successor). The
existing `pty` and `tmux` drivers are reference implementations of the
same surface. The `mode` parameter on the relevant MCP tools is the
user-facing dispatch — extend its enum and add a case in the server's
`dispatchByMode` helper.

### Adding a new render backend

`internal/render` is currently VHS-specific but the public function is
`Render(req RenderReq) (RenderResult, error)`. A second backend (asciinema
+ agg, raw screenshot via headless emulator, etc.) would live in a sibling
package and be selected via a future `backend` parameter on
`tui_render_image`.

### Adding a framework plugin (Ratatui, Textual, Ink)

The teatest path is Bubble Tea-specific by design. v2 candidates would
mirror its structure under `internal/<framework>/`:

- a code generator that produces an in-process harness in the target
  framework's idiom,
- a builder that compiles or otherwise prepares the harness,
- a runner that drives the harness and returns the same shape as
  `tui_test_golden` / `tui_inspect_model`.

The MCP tool surface stays stable — the user just points it at a non-Go
package via the existing `package_path` parameter.

### Adding a new MCP tool

Tools live in `internal/server/` as small registration functions, each
returning the typed handler closure. Add the tool, register it from
`New()`, document it in `docs/mcp-tools.md`, and add a `### Added` entry
to `CHANGELOG.md` under `[Unreleased]`.

## See also

- [`plan.md`](./plan.md) — master plan, phase breakdown, decisions log
- [`mcp-tools.md`](./mcp-tools.md) — full tool reference
- [`white-box-pattern.md`](./white-box-pattern.md) — the `TeaEyesNewModel` convention
- [`compat-ggprompts.md`](./compat-ggprompts.md) — composition with the GGPrompts/TFE skill
- [`workflow.md`](./workflow.md) — start-to-finish narrative tutorial
- [`roadmap.md`](./roadmap.md) — what's coming after v0.1.0
