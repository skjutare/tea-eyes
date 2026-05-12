# Phase 1 — MCP Server Skeleton + pty Driver

**Goal:** A working MCP server with one real tool: `tui_capture_text`. Proves the architecture end-to-end before adding complexity. From a fresh Claude Code session, "show me what `examples/hello-tui` renders" should work.

**Estimated effort:** 1–2 days

**Prerequisites:** Phase 0 complete and pushed.

---

## Prompt to give Claude Code

> Implement Phase 1 of tea-eyes per `docs/plan.md`. This phase delivers a working
> MCP server with one production-quality tool (`tui_capture_text`), a tiny
> Bubble Tea example app for testing, and an end-to-end integration test.
>
> ## Dependencies to add
>
> Add these to `go.mod` (use latest stable versions; pick the canonical import
> paths — verify on pkg.go.dev before adding):
>
> - `github.com/mark3labs/mcp-go` — MCP server framework
> - `github.com/creack/pty` — pty allocation and control
> - `github.com/charmbracelet/x/ansi` — ANSI sequence parsing
>   (alternatively `github.com/hinshun/vt10x` if `x/ansi` lacks a screen-buffer
>   abstraction; document the choice in `internal/capture/doc.go`)
> - `github.com/charmbracelet/bubbletea` — for the example app
> - `github.com/charmbracelet/lipgloss` — for the example app
> - `github.com/stretchr/testify` — assertions in tests
>
> ## Implementation order
>
> Build in this order so each step is independently testable:
>
> ### 1. `internal/keys/`
>
> Implement a parser that turns human-readable key strings into byte sequences
> sent to the pty. Match the notation used by jmlago's tmux skill for
> consistency:
>
> - Plain runes: `"a"`, `"7"`, `"/"` → the literal byte(s)
> - Special keys: `"enter"`, `"escape"`, `"space"`, `"tab"`, `"backspace"`,
>   `"up"`, `"down"`, `"left"`, `"right"`, `"home"`, `"end"`, `"pgup"`,
>   `"pgdown"`, `"delete"`, `"insert"`
> - Function keys: `"f1"` through `"f12"`
> - Modifiers: `"ctrl+c"`, `"alt+x"`, `"ctrl+alt+l"`, `"shift+tab"`
> - Multi-character literals: `"hello world"` → each rune sent in sequence
>
> Export:
>
> ```go
> func Parse(s string) ([]byte, error)
> func ParseSequence(keys []string) ([]byte, error)
> ```
>
> Comprehensive unit tests covering every special key, every modifier
> combination, and edge cases (empty string → error, unknown key → error,
> Unicode runes work).
>
> ### 2. `internal/pty/`
>
> Wrap `creack/pty` to spawn a TUI binary at a fixed terminal size, write
> keystrokes, and read back the raw output buffer over a configurable settle
> period.
>
> ```go
> type Driver struct { /* ... */ }
>
> type SpawnOpts struct {
>     Command   string
>     Args      []string
>     Width     int           // default 80
>     Height    int           // default 24
>     Env       []string      // extended env, defaults to os.Environ()
>     SettleMs  int           // default 300
> }
>
> func New() *Driver
> func (d *Driver) Capture(ctx context.Context, opts SpawnOpts, keys [][]byte) ([]byte, error)
> ```
>
> The Capture flow:
> 1. Spawn the process attached to a pty sized W×H (use `pty.Setsize`).
> 2. Wait `SettleMs` for the initial render.
> 3. For each key sequence: write to pty, wait `SettleMs`.
> 4. Read everything available from the pty.
> 5. Send SIGTERM, wait briefly, SIGKILL if needed.
> 6. Return the accumulated raw output bytes.
>
> Important: handle the read side in a goroutine with a buffered channel so the
> child process doesn't block on a full pty buffer. Set `TERM=xterm-256color`
> in the child env unless the user overrode it.
>
> ### 3. `internal/capture/`
>
> Take raw pty output bytes and produce a clean text representation of the
> visible cell grid.
>
> ```go
> type Frame struct {
>     Width  int
>     Height int
>     Text   string  // newline-separated lines, no trailing newline
> }
>
> func RenderFrame(raw []byte, width, height int, stripANSI bool) (Frame, error)
> ```
>
> Use the chosen ANSI library to feed the bytes through a virtual terminal,
> then dump the visible buffer. When `stripANSI` is false, preserve SGR
> sequences in the output so Claude can see colors if needed.
>
> Tests: feed in known sequences (a "hello world" with a styled border) and
> assert the output matches expected lines.
>
> ### 4. `internal/server/` + `cmd/tea-eyes/main.go`
>
> Stdio MCP server using `mark3labs/mcp-go`. Register one tool:
>
> **Tool: `tui_capture_text`**
>
> Inputs:
> | name | type | required | default | description |
> |------|------|----------|---------|-------------|
> | `command` | string | yes | — | binary or shell command to run |
> | `args` | []string | no | `[]` | arguments to pass |
> | `keys` | []string | no | `[]` | sequence of key inputs to send |
> | `width` | int | no | 80 | terminal width in columns |
> | `height` | int | no | 24 | terminal height in rows |
> | `settle_ms` | int | no | 300 | wait between key sends |
> | `strip_ansi` | bool | no | true | strip SGR codes from output |
> | `cwd` | string | no | "" | working directory (default: server cwd) |
>
> Output:
>
> ```json
> {
>   "text": "...",
>   "width": 80,
>   "height": 24,
>   "raw_bytes": 1234
> }
> ```
>
> Errors must be returned as MCP tool errors with helpful messages (e.g.
> "command not found: foo", "unknown key 'ctrl+meow'", "process exited with
> status N before settle").
>
> `main.go` should be small: parse flags (`--version`, `--log-file`), set up
> logging (default to stderr), construct the server, register the tool, run on
> stdio.
>
> ### 5. `examples/hello-tui/`
>
> A minimal Bubble Tea app:
> - Shows a centered greeting "Hello, tea-eyes!" inside a Lipgloss border.
> - Pressing `q` or `ctrl+c` quits.
> - Pressing space toggles a counter line.
>
> Include its own `go.mod` (so it doesn't pollute the parent module's
> dependencies — use a `replace` directive only if needed for testing).
> Actually: since the parent module already needs bubbletea and lipgloss for
> tests, the simplest approach is to make hello-tui a `package main` under the
> parent module. Do that.
>
> ### 6. `test/integration/capture_test.go`
>
> An end-to-end test that:
> 1. Builds `cmd/tea-eyes` to a temp binary.
> 2. Builds `examples/hello-tui` to a temp binary.
> 3. Spawns the MCP server, connects an in-process MCP client, calls
>    `tui_capture_text` with the hello-tui binary as `command`.
> 4. Asserts the returned text contains "Hello, tea-eyes!".
> 5. Calls again with `keys: [" "]` and asserts the counter line appears.
> 6. Calls again with `keys: ["q"]` and asserts the process exited cleanly.
>
> Make this test runnable as `make test` (it's slow-ish, ~5s, that's fine).
>
> ## Documentation updates
>
> - Update `docs/mcp-tools.md` with a full reference for `tui_capture_text`,
>   including the input/output table above and 2–3 example invocations.
> - Update `README.md` "Quickstart" section with: install, build, run a
>   capture against `examples/hello-tui` from the command line using a
>   one-shot MCP client invocation (use `mcp-go`'s built-in test client or
>   a tiny shell snippet).
> - Update `CHANGELOG.md` under `## [Unreleased]`:
>   - `### Added` — `tui_capture_text` tool, pty driver, key parser,
>     hello-tui example.
>
> ## Acceptance criteria
>
> 1. `make build test lint` all pass.
> 2. `go test -race ./...` passes (no data races).
> 3. The integration test reliably passes 10 times in a row
>    (`go test -count=10 ./test/integration/...`).
> 4. Manually: connect Claude Code to the locally built `tea-eyes` binary as
>    an MCP server (document the exact `claude mcp add` command in README) and
>    ask it to capture `examples/hello-tui`. Claude should see the greeting.
> 5. CI pipeline goes green on push.
>
> ## Anti-scope (do NOT do in this phase)
>
> - VHS / image rendering — Phase 2
> - teatest integration — Phase 3
> - Skills or subagents — Phase 4 / 5
> - tmux driver — Phase 6
> - Any caching, parallelism, or performance work
>
> When done, summarize what you built, confirm all acceptance criteria, and
> stop. Phase 2 is a separate prompt.
