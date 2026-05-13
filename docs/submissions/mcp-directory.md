# Submission draft — mcp.directory

**Listing type:** MCP server

**Name:** tea-eyes

**Tagline:** Playwright for terminals.

**Description:**

tea-eyes is an MCP server that lets an LLM client (Claude Code primarily)
*see* what a terminal user interface renders. Three complementary
strategies: text capture via pty/tmux, true-pixel image via Charm's VHS,
and in-process introspection of Bubble Tea models via teatest.

Shipped as a Claude Code plugin that bundles two reference subagents
(`tui-designer`, `tui-tester`) and two skills (`tea-eyes-loop`,
`tea-eyes-bubbletea`) alongside the MCP server itself.

**Install:**

```sh
go install gitlab.com/skjutare/tea-eyes/cmd/tea-eyes@latest
```

**Run:**

```sh
tea-eyes serve   # stdio MCP server
```

**Tools:** 5 (see README / docs/mcp-tools.md).

**Repository:** https://gitlab.com/skjutare/tea-eyes

**License:** MIT

**Author:** Christoffer Skjutare

**Demo:** https://gitlab.com/skjutare/tea-eyes/-/blob/main/docs/img/multi-pane-demo.png
