# Submission draft — mcpmarket.com

**Listing type:** MCP server (with skill + agent bundle)

**Name:** tea-eyes

**Tagline:** Playwright for terminals — visual feedback for TUI development
with Claude Code.

**Short description (≤200 chars):**

> An MCP server that lets Claude *see* a terminal user interface three
> ways: text capture, true-pixel image render, or in-process Bubble Tea
> model state. Ships with two subagents and two skills.

**Long description:**

tea-eyes closes the visual feedback gap that has historically made TUI
development with Claude Code feel blind compared to web development with
Playwright.

It exposes five MCP tools:

- `tui_capture_text` — fast ASCII grid of any TUI binary via pty or tmux
- `tui_render_image` — true-pixel PNG/GIF via Charm's VHS, returned as an
  MCP image content block
- `tui_test_golden` — in-process Bubble Tea golden-file tests via teatest
- `tui_inspect_model` — JSON dump of a Bubble Tea model's exported fields
- `tui_session_attach_hint` — `tmux attach` command for watching Claude
  work live

Two reference subagents (`tui-designer`, `tui-tester`) and two skills
(`tea-eyes-loop`, `tea-eyes-bubbletea`) ship in the same plugin. The
Bubble Tea skill is intentionally additive to the GGPrompts/TFE bubbletea
skill rather than a replacement.

**Install:**

```sh
claude plugin install gitlab.com/skjutare/tea-eyes
# or for just the MCP server:
go install gitlab.com/skjutare/tea-eyes/cmd/tea-eyes@latest
claude mcp add tea-eyes -- tea-eyes serve
```

**Repository:** https://gitlab.com/skjutare/tea-eyes
(mirror: https://github.com/skjutare/tea-eyes)

**License:** MIT

**Author:** Christoffer Skjutare

**Demo:** see README banner image (and the GIF once recorded)

**Tags / keywords:** tui, bubble-tea, terminal, lipgloss, vhs, teatest,
testing, golden-files, go, playwright, charm
