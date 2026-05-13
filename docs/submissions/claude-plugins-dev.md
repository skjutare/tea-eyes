# Submission draft — claude-plugins.dev

**Listing type:** Claude Code plugin

**Name:** tea-eyes

**Tagline:** Playwright for terminals — visual feedback for TUI
development.

**What's in the bundle:**

- 1 MCP server (`tea-eyes serve`) with 5 tools
- 2 reference subagents (`tui-designer`, `tui-tester`)
- 2 skills (`tea-eyes-loop`, `tea-eyes-bubbletea`)

**Why it's interesting on this directory:**

This is one of the rare plugins that bundles **all three** Claude Code
extension surfaces — MCP server, skills, and subagents — into one focused
workflow. The subagents bake the render → look → edit → re-render loop
into a tool list, the skills teach Claude when to pick which tool, and
the MCP server provides the underlying capabilities.

The Bubble Tea skill is deliberately **additive** to the existing
GGPrompts/TFE bubbletea skill (which encodes the "4 Golden Rules" of
Bubble Tea layout). Install both; tea-eyes verifies via real renders
while GGPrompts prescribes what good looks like.

**Install:**

```sh
claude plugin install gitlab.com/skjutare/tea-eyes
```

Then `tea-eyes doctor` to verify the optional external dependencies
(`vhs`, `ttyd`, `ffmpeg`, `tmux`).

**Repository:** https://gitlab.com/skjutare/tea-eyes
(mirror: https://github.com/skjutare/tea-eyes)

**License:** MIT

**Author:** Christoffer Skjutare

**Categories:** developer-tools, testing, ui-development
