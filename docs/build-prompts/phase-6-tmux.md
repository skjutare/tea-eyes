# Phase 6 — tmux Driver (v2 Process Model)

**Goal:** Add tmux as an alternative driver for interactive sessions, alongside the existing pty driver. All capture/render/test tools gain an optional `mode` parameter. tmux mode lets the user *watch* Claude work in another terminal — same workflow as the jmlago skill, but exposed through the MCP for first-class composition.

**Estimated effort:** 1 day

**Prerequisites:** Phases 0–5 complete. tmux installed locally for development.

---

## Prompt to give Claude Code

> Implement Phase 6 of tea-eyes per `docs/plan.md`. Add a tmux-based process
> driver as an alternative to the pty driver. The user can choose per-call
> which mode to use; defaults stay on pty for backward compatibility.
>
> ## Why tmux mode
>
> - The user can attach to the tmux session in another terminal and watch
>   Claude drive the TUI live — invaluable for trust-building and debugging.
> - Persistent sessions: the TUI keeps running between tool calls, so you can
>   inspect intermediate state without re-spawning.
> - Compatible with apps that misbehave under non-tmux ptys.
>
> ## Why pty stays the default
>
> - No external dependency.
> - Faster startup (no session management).
> - Better for CI / non-interactive use.
> - Cleaner teardown.
>
> ## Implementation
>
> ### 1. `internal/tmux/`
>
> Wrap the `tmux` CLI. Don't shell out for every byte — use longer-lived
> commands where possible. Reference the jmlago skill for the canonical
> command vocabulary.
>
> ```go
> type Driver struct {
>     tmuxBin string  // path to tmux binary; default "tmux"
> }
>
> type SessionOpts struct {
>     Name        string  // session name; if empty, generate "teaeyes-<random>"
>     Width       int     // initial pane width
>     Height      int     // initial pane height
>     Detached    bool    // create detached (always true for tea-eyes use)
> }
>
> type Session struct {
>     Name string
>     // ...
> }
>
> func New() *Driver
> func (d *Driver) NewSession(opts SessionOpts) (*Session, error)
> func (d *Driver) AttachExisting(name string) (*Session, error)
> func (d *Driver) KillSession(s *Session) error
>
> func (s *Session) RunCommand(cmd string) error                  // sends a command + Enter
> func (s *Session) SendKeys(keys [][]byte) error                 // raw key bytes
> func (s *Session) CapturePane(includeColors bool) (string, error)
> func (s *Session) Resize(width, height int) error
> func (s *Session) PaneSize() (width, height int, err error)
> ```
>
> Notes:
> - Use `tmux new-session -d -s <name> -x <w> -y <h>` to create.
> - Use `tmux send-keys -t <name>` to send input. For raw bytes, you may need
>   to encode them; mirror jmlago's notation by translating from the keys
>   package's parsed form.
> - Use `tmux capture-pane -t <name> -p` (add `-e` for ANSI colors).
> - Be conservative about when you kill the session. If the user passed an
>   existing session name, never kill it on cleanup. Only kill sessions
>   tea-eyes itself created.
>
> ### 2. Driver abstraction
>
> Refactor existing code so that pty and tmux share a common interface:
>
> ```go
> // in internal/driver (new package)
> type Mode string
>
> const (
>     ModePTY  Mode = "pty"
>     ModeTmux Mode = "tmux"
> )
>
> type Driver interface {
>     Capture(ctx context.Context, opts CaptureOpts) ([]byte, error)
> }
>
> type CaptureOpts struct {
>     Command   string
>     Args      []string
>     Keys      [][]byte
>     Width     int
>     Height    int
>     SettleMs  int
>     Cwd       string
>     // tmux-only:
>     SessionName string  // if set, use this session; if not, create one
>     Persist     bool    // if true, don't kill the session on completion
> }
>
> func NewDriver(mode Mode) (Driver, error)
> ```
>
> Existing pty-using code moves behind this interface. Existing callers update
> to pick the mode.
>
> ### 3. MCP tool updates
>
> Every tool gains an optional `mode` input:
>
> | name | type | required | default | description |
> |------|------|----------|---------|-------------|
> | `mode` | enum | no | "pty" | "pty" or "tmux" |
> | `tmux_session` | string | no | "" | tmux mode only: existing session name; if empty, ephemeral session is created and killed after the call |
> | `tmux_persist` | bool | no | false | tmux mode only: keep the session alive after the call (implies `tmux_session` will be returned) |
>
> Tool output gains (when applicable):
>
> ```json
> {
>   "tmux_session": "teaeyes-x8a2"  // present if mode=tmux and persist=true
> }
> ```
>
> This applies to `tui_capture_text` and (with caveats) `tui_render_image`.
> For `tui_render_image` in tmux mode: VHS doesn't natively render *from* a
> tmux session. Two options:
>
> - **Option A** (simpler, recommended for v1): document that
>   `tui_render_image` always uses pty mode internally, regardless of `mode`
>   parameter. Reject `mode=tmux` with a clear error.
> - **Option B** (more work): use VHS to record a session that connects to
>   the tmux session via `tmux attach`. Probably not worth the complexity.
>
> Go with A. Document the limitation.
>
> For `tui_test_golden` and `tui_inspect_model`: these are in-process
> teatest-based and don't use a driver at all. The `mode` parameter is
> rejected if provided.
>
> ### 4. New helper tool: `tui_session_attach_hint`
>
> A small tool that returns the shell command the user can run to attach to
> a persistent tmux session, e.g. `tmux attach -t teaeyes-x8a2`. This makes
> it trivial for Claude to tell the user "run `<command>` to watch."
>
> Inputs: `session_name` (string, required).
> Output: `{ "command": "tmux attach -t teaeyes-x8a2", "exists": true }`.
>
> ### 5. Doctor update
>
> Add tmux to `tea-eyes doctor` output. Report version. If missing, suggest
> install (`brew install tmux` / package-manager hint). Note that tmux is
> *optional*, not required.
>
> ### 6. Tests
>
> - Unit tests for `internal/tmux/` against a real tmux (skip if tmux not on
>   PATH).
> - Integration test in `test/integration/tmux_test.go`:
>   1. Skip if no tmux.
>   2. Capture `examples/hello-tui` via `mode=tmux`. Assert greeting present.
>   3. Capture again with `tmux_persist: true`, get session name back.
>   4. Capture a third time with that session name; assert state is preserved.
>   5. Clean up the session.
> - Verify behavior parity with pty mode for the basic capture: the captured
>   text should be substantially the same (allow minor whitespace/trailing
>   line differences; assert on substring presence).
>
> ## Documentation updates
>
> - `docs/mcp-tools.md`: update every applicable tool with the new `mode`,
>   `tmux_session`, `tmux_persist` parameters. Add a section "Choosing a
>   mode" with the tradeoffs.
> - `docs/workflow.md`: new section "Watching Claude drive the TUI" — covers
>   the `mode=tmux` + `tmux_persist=true` pattern, with a worked example
>   showing the user attaching to the session in another terminal.
> - `README.md`: brief mention in Quickstart that pty is default; tmux is
>   opt-in. Link to the workflow doc.
> - `CHANGELOG.md`: `### Added` — tmux driver, `mode` parameter on tools,
>   `tui_session_attach_hint` tool, doctor reports tmux.
>
> ## Acceptance criteria
>
> 1. All previous tests still pass.
> 2. The new tmux integration test passes locally with tmux installed.
> 3. CI: tmux integration tests are tagged and gated; pty tests still run
>    without tmux.
> 4. Manually: ask Claude to capture a TUI in tmux mode with persist=true,
>    then attach to that session in a separate terminal and verify it's
>    actually showing the TUI.
> 5. `tea-eyes doctor` reports tmux availability.
> 6. Attempting `tui_render_image` with `mode=tmux` returns a clear,
>    actionable error.
>
> ## Anti-scope
>
> - VHS-from-tmux rendering — explicitly out of scope for v1.
> - Multi-pane / multi-window tmux orchestration — single-pane sessions only.
> - Cross-host tmux (over SSH) — not supported, not tested.
>
> When done, summarize, confirm acceptance, stop.
