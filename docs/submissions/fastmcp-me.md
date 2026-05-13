# Submission draft — fastmcp.me

**Listing type:** Agent skill collection

**Name:** tea-eyes

**Tagline:** Visual feedback loop for TUI development — render, look,
edit, re-render.

**Skills:**

- `tea-eyes-loop` (framework-agnostic) — teaches the capture →
  reason → edit → re-capture loop and when to pick which tool.
- `tea-eyes-bubbletea` (Bubble Tea specific) — covers the
  `TeaEyesNewModel` white-box pattern, color-profile discipline for
  stable goldens, and the teatest workflow. **Strictly additive** to the
  GGPrompts/TFE bubbletea skill (which encodes layout rules).

**Backing MCP server:**

Both skills assume the tea-eyes MCP server is installed and connected;
it provides `tui_capture_text`, `tui_render_image`, `tui_test_golden`,
`tui_inspect_model`, and `tui_session_attach_hint`.

**Install:**

```sh
claude plugin install gitlab.com/skjutare/tea-eyes
```

This installs both skills, both subagents, and the MCP server.

**Repository:** https://gitlab.com/skjutare/tea-eyes

**License:** MIT

**Author:** Christoffer Skjutare
