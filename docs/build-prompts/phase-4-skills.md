# Phase 4 — Companion Skills

**Goal:** Skills that teach Claude how to *use* tea-eyes well, plus the GGPrompts compatibility layer. Two skills: framework-agnostic loop guidance, and Bubble Tea-specific guidance that defers layout rules to the GGPrompts/TFE bubbletea skill.

**Estimated effort:** ½–1 day

**Prerequisites:** Phases 0–3 complete. The MCP server now exposes
`tui_capture_text`, `tui_render_image`, `tui_test_golden`, `tui_inspect_model`.

---

## Prompt to give Claude Code

> Implement Phase 4 of tea-eyes per `docs/plan.md`. Author the two SKILL.md
> files plus the Claude Code plugin manifest. The skills are how Claude
> *discovers and triggers* the right tea-eyes behaviors at the right moments.
>
> ## Reference material to read first
>
> Before writing the skills, fetch and study:
> 1. The GGPrompts/TFE bubbletea skill SKILL.md — the canonical reference for
>    Bubble Tea layout rules. URL:
>    https://github.com/GGPrompts/TFE/tree/main/.claude/skills/bubbletea
> 2. The jmlago "Debug TUIs with tmux" skill — the workflow this skill
>    formalizes. URL: https://jmlago.github.io/skills/debug-tuis-with-tmux.md
> 3. Anthropic's skill authoring guide and the Claude Code plugin docs for
>    the `plugin.json` schema (search docs.claude.com).
>
> Goal: tea-eyes-bubbletea must be **additive**, not redundant. If both skills
> are installed, they should compose cleanly with no conflicting rules.
>
> ## Skill 1: `plugin/skills/tea-eyes-loop/SKILL.md`
>
> Framework-agnostic. Teaches the visual feedback loop and tool selection.
>
> Frontmatter:
>
> ```yaml
> ---
> name: tea-eyes-loop
> description: |
>   Use this skill when developing, modifying, or debugging any terminal user
>   interface (TUI) application — Bubble Tea, Ratatui, Textual, Ink, or
>   anything else that renders to a terminal. Teaches Claude how to use the
>   tea-eyes MCP tools (tui_capture_text, tui_render_image, tui_test_golden,
>   tui_inspect_model) to see what the TUI actually renders, judge layout and
>   visual quality, and iterate. Trigger any time you would otherwise be
>   guessing what the TUI looks like — e.g. when changing layout, colors,
>   spacing, key bindings, or any visual element.
> license: MIT
> ---
> ```
>
> Body sections:
>
> 1. **The loop**: capture → reason → edit → re-capture. State explicitly that
>    Claude must verify visual changes by re-capturing, not by inspecting code
>    alone.
>
> 2. **Which tool when** — a decision table:
>    | Goal | Tool | Why |
>    |------|------|-----|
>    | Quick structural check ("is the panel where I expect?") | `tui_capture_text` | Fast, cheap, sufficient for ASCII layout |
>    | Color/spacing/font/typography judgment | `tui_render_image` | Need pixels, not cells |
>    | Lock in a behavior so it can't regress | `tui_test_golden` | Only Bubble Tea; persistent assertion |
>    | Inspect runtime state during debugging | `tui_inspect_model` | Only Bubble Tea; structured model dump |
>
> 3. **Anti-patterns** (with examples of each):
>    - Calling `tui_render_image` on every iteration when text capture would
>      do — wastes time.
>    - Skipping `tui_test_golden` after locking in a behavior that matters.
>    - Forgetting to set fixed `width` and `height` — non-deterministic.
>    - `settle_ms` too short for animated/async TUIs — flaky captures.
>    - Theme drift breaking golden files — pin `color_profile: "Ascii"` for
>      behavior tests.
>
> 4. **The standard loop** (a worked example showing 4–5 turns of iteration on
>    a hypothetical layout change).
>
> 5. **Composing with other skills**: explicitly mention that if a
>    framework-specific skill is also installed (e.g. tea-eyes-bubbletea,
>    GGPrompts/TFE bubbletea), follow that skill's design rules first.
>    tea-eyes-loop is about the *feedback mechanism*, not the design rules.
>
> ## Skill 2: `plugin/skills/tea-eyes-bubbletea/SKILL.md`
>
> Bubble Tea-specific guidance. Strictly additive to GGPrompts/TFE.
>
> Frontmatter:
>
> ```yaml
> ---
> name: tea-eyes-bubbletea
> description: |
>   Use this skill when building or modifying Bubble Tea TUI applications in Go
>   with the tea-eyes MCP tools available. Adds Bubble Tea-specific guidance to
>   the tea-eyes-loop skill: the white-box TeaEyesNewModel pattern for
>   in-process testing, color profile guidance for stable golden files, the
>   teatest workflow, and image rendering for visual review. This skill is
>   ADDITIVE to the GGPrompts/TFE bubbletea skill (which encodes layout rules)
>   — if both are installed, defer all layout/component design rules to that
>   skill and use this one only for the tea-eyes-specific workflow.
> license: MIT
> ---
> ```
>
> Body sections:
>
> 1. **Composition statement** (top of body, prominent):
>    > This skill is additive to the GGPrompts/TFE bubbletea skill. If that
>    > skill is installed, follow its 4 Golden Rules for layout. This skill
>    > does NOT duplicate or override those rules. It only adds the tea-eyes
>    > workflow on top.
>
> 2. **The white-box pattern**: how to add `TeaEyesNewModel()` to the user's
>    package under the `teaeyes` build tag. Code example. When the model
>    constructor takes arguments, expose a default-construct wrapper.
>
> 3. **Color profile discipline**:
>    - Behavior tests: `color_profile: "Ascii"` to decouple from theme changes.
>    - Color tests: `color_profile: "TrueColor"` and assert on specific SGR
>      bytes (e.g. `bytes.Contains(out, []byte("48;2;"))`), not on RGB values.
>    - Never mix the two in one golden file.
>
> 4. **The teatest workflow**:
>    - First time: omit `update_golden`, get a "no golden" creation.
>    - Subsequent: run with no flag, expect match.
>    - Intentional change: re-run with `update_golden: true`, review diff in
>      git.
>
> 5. **Image rendering for design**: render `examples/multi-pane`-style apps
>    with multiple themes by varying `theme` to compare aesthetics quickly.
>
> 6. **Common Bubble Tea gotchas tea-eyes catches**:
>    - Initial render doesn't match later renders (forgot `Init()` cmd).
>    - `tea.WindowSizeMsg` not handled → `width`/`height` ignored.
>    - Async cmds need longer `settle_ms`.
>
> 7. **Pointer back to GGPrompts**: end with "for layout, border, mouse, and
>    component design rules, see the GGPrompts/TFE bubbletea skill."
>
> ## Plugin manifest: `plugin/plugin.json`
>
> Per Claude Code plugin schema, declare:
> - Plugin name: `tea-eyes`
> - Version: read from `VERSION` file or `git describe`
> - Description: one line.
> - Author: `Christoffer Skjutare`
> - License: MIT
> - MCP server: stdio command `tea-eyes serve` (add a `serve` subcommand to
>   `cmd/tea-eyes/main.go` if not already present — it should default to
>   serve for backward compat).
> - Skills: both SKILL.md paths.
> - Subagents: both `plugin/agents/*.md` paths (will be created in Phase 5;
>   reference them now and Phase 5 will fill them in).
>
> Verify the schema by reading the latest plugin docs at
> https://docs.claude.com (search for "Claude Code plugins").
>
> ## `docs/compat-ggprompts.md`
>
> A short doc explaining the composition:
> - Install both skills.
> - GGPrompts handles "what should the layout look like."
> - tea-eyes handles "is the layout actually rendering that way."
> - Suggested install order and a quickstart that uses both together.
> - Link to GGPrompts skill repo.
>
> ## Acceptance criteria
>
> 1. Both SKILL.md files lint clean (no broken markdown, frontmatter parses).
> 2. The `description` fields trigger reliably — test by reading them aloud
>    and asking yourself "would Claude know when to use this?". Iterate until
>    yes.
> 3. `plugin/plugin.json` validates against the Claude Code plugin schema
>    (run any provided validator, or hand-check against the docs).
> 4. `docs/compat-ggprompts.md` exists and is accurate.
> 5. Manually: in a Claude Code session with the plugin loaded, mention "I
>    want to change the panel border color in my Bubble Tea app." Claude
>    should pull in tea-eyes-bubbletea (and GGPrompts if installed) and start
>    the loop with `tui_render_image`.
> 6. CHANGELOG updated: `### Added` — tea-eyes-loop skill, tea-eyes-bubbletea
>    skill, plugin manifest, GGPrompts compatibility doc.
>
> ## Anti-scope
>
> - Subagents — Phase 5
> - Re-implementing layout rules from GGPrompts — never; defer.
> - Adding skills for other frameworks (Ratatui, Textual) — out of v1; mention
>   in roadmap only.
>
> When done, summarize, confirm acceptance, stop.
