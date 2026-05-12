# tea-eyes

> Visual feedback for TUI development with Claude Code.

**Status:** alpha — under construction.

## What this is

tea-eyes is a Claude Code plugin — an MCP server plus companion skills and
reference subagents — that lets Claude Code *see* what it's building when
developing terminal user interfaces. It exposes three complementary
capture/render strategies:

- **Text capture** via pty (and tmux, later) + a VT emulator — fast ASCII
  grid of any TUI binary.
- **Image rendering** via [VHS](https://github.com/charmbracelet/vhs) — true-
  pixel PNG/GIF for design review.
- **In-process introspection** via teatest — structured golden output and
  model-state dumps for Bubble Tea apps specifically.

Claude picks the right strategy per task, guided by a packaged skill.

## What this is not

tea-eyes is **not** a replacement for the GGPrompts/TFE bubbletea Claude
Code skill, which encodes Bubble Tea layout best practices (the "4 Golden
Rules"). tea-eyes is *additive* — install both for the full design loop.
See `docs/compat-ggprompts.md` (populated in Phase 4) for details.

## Why

Hatchet's "Building a TUI is easy now" post validated the workflow: Claude
Code + tmux + Bubble Tea is genuinely productive. Skills like jmlago's
"Debug TUIs with tmux" already package an ad-hoc version of this loop.
tea-eyes formalizes the same pattern as a first-class MCP server with a
typed tool surface and three capture strategies instead of one.

## Status / Roadmap

- [x] Phase 0 — Bootstrap (repo skeleton, license, CI)
- [ ] Phase 1 — MCP server + pty driver + `tui_capture_text`
- [ ] Phase 2 — VHS image rendering + `tui_render_image`
- [ ] Phase 3 — teatest harness + `tui_test_golden` + `tui_inspect_model`
- [ ] Phase 4 — Skills + plugin manifest
- [ ] Phase 5 — Reference subagents (`tui-designer`, `tui-tester`)
- [ ] Phase 6 — tmux driver + `mode` parameter
- [ ] Phase 7 — Release polish + v0.1.0

Full plan: [`docs/plan.md`](docs/plan.md). Per-phase prompts:
[`docs/build-prompts/`](docs/build-prompts/).

## Prior art and credits

- [Charm](https://charm.sh) — Bubble Tea, Lipgloss, Bubbles, VHS, teatest
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) — Go MCP framework
- GGPrompts / TFE bubbletea Claude Code skill
- jmlago "Debug TUIs with tmux" Claude Code skill
- rigerc/bubbletea-v2-scaffold
- Hatchet — "Building a TUI is easy now"

See [`NOTICE`](NOTICE) for the full attribution.

## License

MIT. See [`LICENSE`](LICENSE) and [`NOTICE`](NOTICE).
