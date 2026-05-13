# Subagents

tea-eyes ships two reference subagents — focused, narrow assistants that
combine the MCP tools and skills into a single workflow. Each is
defined under [`plugin/agents/`](../plugin/agents/) and registered in
[`plugin/plugin.json`](../plugin/plugin.json).

Subagents are the highest-level entry point into tea-eyes. You invoke
them by name (or let Claude Code auto-select via their `description`
field) and they take over the conversation with a tight system prompt
and a curated tool list.

## `tui-designer`

**When to invoke.** Visual work on a TUI: layout, spacing, borders,
colors, focus states, theming, typography — anything where "what does
it look like" matters more than "does it work."

**Example prompts that should trigger it.**

- "Tweak the focus color of the right panel in `examples/multi-pane` to
  be more vibrant."
- "The status bar feels cramped. Add a bit of breathing room and a
  subtle border."
- "Make the title block stand out more."

**Tools it has.** `Read`, `Write`, `Edit`, `Bash`, `Glob`, `Grep`,
`mcp__tea-eyes__tui_capture_text`, `mcp__tea-eyes__tui_render_image`.

**Tools it deliberately does *not* have.** The teatest tools
(`tui_test_golden`, `tui_inspect_model`). Design iteration is about
rendered pixels, not golden files — keeping the tool list narrow keeps
the agent focused on its job.

**How it composes with the skills.** When the `tea-eyes-bubbletea` or
GGPrompts/TFE bubbletea skills are loaded, `tui-designer` defers to
them for layout rules (the 4 Golden Rules etc.) before exercising
creative judgment. The skills encode *what good looks like*; the agent
verifies *whether the implementation matches* via real renders.

## `tui-tester`

**When to invoke.** Locking in Bubble Tea TUI behavior: adding golden
tests, updating goldens after an intentional change, triaging a flaky
test, or bootstrapping the `TeaEyesNewModel` white-box hook in a
package that doesn't have one yet.

**Example prompts that should trigger it.**

- "Add a golden test that locks in the current state of
  `examples/hello-tui` after pressing space."
- "The right-panel focus behavior changed — update the golden to
  match."
- "`TestNav_filter` is flaky in CI, can you stabilize it?"

**Tools it has.** `Read`, `Write`, `Edit`, `Bash`, `Glob`, `Grep`,
`mcp__tea-eyes__tui_test_golden`, `mcp__tea-eyes__tui_inspect_model`,
`mcp__tea-eyes__tui_capture_text`.

**Tools it deliberately does *not* have.** `tui_render_image`.
Behavior tests must be terminal-environment-independent; image renders
are explicitly off-limits in a testing context to remove the temptation
to "verify by eye" instead of by golden diff.

**How it composes with the skills.** The `tea-eyes-bubbletea` skill
covers the `TeaEyesNewModel` pattern and the color-profile discipline.
`tui-tester` enforces both as non-negotiable: behavior tests use the
ASCII profile, color tests use TrueColor and assert on SGR markers,
and every tested package gets a build-tagged `TeaEyesNewModel()` hook.

## Composition with the rest of the plugin

```
User prompt
     │
     ▼
 tui-designer  /  tui-tester           ← subagents: workflow + persona
     │                │
     ▼                ▼
 tea-eyes-loop  /  tea-eyes-bubbletea  ← skills: rules + how-to
     │                │
     ▼                ▼
 mcp__tea-eyes__tui_*                  ← MCP tools: capabilities
     │
     ▼
 pty / VHS / teatest                   ← drivers
```

The subagents are intentionally narrow. There is **no** general-purpose
"tui-engineer" agent — that role belongs to Claude Code itself, with
the skills loaded.
