# Roadmap

Indicative, not commitments. tea-eyes is a side project on a best-effort
SLA (see [`CONTRIBUTING.md`](../CONTRIBUTING.md)). Items move in and out
of this list as real use surfaces real needs.

## v0.1.0 — shipped

Phases 0–7 from the master [`plan.md`](./plan.md):

- Framework-agnostic text capture (`tui_capture_text`) via pty or tmux
- Image rendering (`tui_render_image`) via VHS with content-addressed caching
- Bubble Tea in-process teatest harness with `tui_test_golden` and
  `tui_inspect_model`
- Two reference subagents (`tui-designer`, `tui-tester`)
- Two skills (`tea-eyes-loop`, `tea-eyes-bubbletea`)
- `tui_session_attach_hint` for watching Claude drive a tmux session live
- Cross-platform release artifacts (linux/darwin/windows × amd64/arm64)
- Distribution as a Claude Code plugin

## v0.2 — framework plugins

Lift the teatest-style in-process pattern to a second TUI framework. Most
likely candidates, in rough priority order:

- **Ratatui** (Rust) — large user base, mature ecosystem, good fit for
  `tui_test_golden` analog
- **Textual** (Python) — has a built-in `Pilot` driver that aligns with
  the teatest model; could expose it through tea-eyes
- **Ink** (JS/TS) — React-based; testing story already exists via
  `ink-testing-library`

Each framework plugin keeps the MCP tool surface identical (`package_path`
gains a different meaning) and lives in `internal/<framework>/`. See
[`architecture.md`](./architecture.md) §Extension points.

## v0.3 — visual diffing

`tui_render_image` cold renders are 2–5 seconds; pairs of them make a
natural diff target. v0.3 adds:

- `tui_diff_image` — render two states, diff pixel-for-pixel with bounding
  boxes around regions of change, return both images plus the diff
- "Did this layout change land?" becomes a single MCP call rather than two
  renders + visual comparison in the agent's head

The bounding-box list also unlocks scripted assertions ("expect the right
panel border region to change, nothing else").

## v0.4 — record & replay

Capture a real user session against a TUI (keystroke stream + final
state), persist it, and replay it as a regression test. Fits between the
ad-hoc `tui_capture_text` loop and the rigorous `tui_test_golden` workflow.

## v1.0 — API stability

Lock the MCP tool surface, the `pkg/teaeyes` Go API, and the plugin
manifest schema. SemVer commitment from v1.0 onward.

Gating criteria (all of):

- ≥3 months of v0.x in active use without major surface churn
- ≥1 external project depending on `pkg/teaeyes` or driving the MCP server
  from a non-Claude-Code MCP client
- Documented deprecation policy for the tool surface

## Beyond v1 — speculative

- Cross-host operation (drive a TUI on a remote host via SSH or remote tmux)
- Multi-pane / multi-window tmux orchestration
- A web dashboard (likely a "maybe" — runs against the existing MCP
  surface, no server-side state required)
- Anthropic official Go MCP SDK migration (revisit when it hits 1.0)

## Won't do, ever

These are permanently out of scope (per [`plan.md`](./plan.md) §2):

- Telemetry or analytics of any kind
- Replacing or duplicating GGPrompts/TFE bubbletea layout rules
- Becoming a TUI framework
- Locking users into a specific TUI library
