# Phase 5 — Reference Subagents

**Goal:** Two ready-to-use subagent definitions that combine the MCP + skills into actual workflows. These are the highest-level "use tea-eyes" entry points — users invoke them by name and get a focused agent with the right tools and the right system prompt.

**Estimated effort:** ½ day

**Prerequisites:** Phases 0–4 complete.

---

## Prompt to give Claude Code

> Implement Phase 5 of tea-eyes per `docs/plan.md`. Author two subagent
> definitions — one for design iteration, one for test maintenance. Subagents
> are the user-facing entry points that tie the MCP tools and skills together
> into focused workflows.
>
> ## Reference material
>
> Read first:
> - https://code.claude.com/docs/en/sub-agents — the canonical format and
>   semantics. Pay close attention to:
>   - Frontmatter fields: `name`, `description`, `tools`, `model`
>   - The body becomes the system prompt; subagents do NOT inherit the main
>     Claude Code system prompt
>   - Subagents start in the main conversation's CWD; `cd` doesn't persist
>     between Bash calls
>   - The `description` field drives automatic invocation
>
> ## Subagent 1: `plugin/agents/tui-designer.md`
>
> Frontmatter:
>
> ```yaml
> ---
> name: tui-designer
> description: |
>   Use this agent to iterate on the visual design of a terminal user
>   interface — layout, spacing, colors, borders, typography, focus states.
>   The agent uses the tea-eyes MCP tools to render the TUI as actual images
>   so it can see and judge the design, then edits the source, then
>   re-renders to verify changes. Invoke when the user asks for visual
>   tweaks, layout work, theme work, or any change where "what does it look
>   like" matters more than "does it work."
> tools: Read, Write, Edit, Bash, Glob, Grep, mcp__tea-eyes__tui_capture_text, mcp__tea-eyes__tui_render_image
> model: inherit
> ---
> ```
>
> System prompt body (the bulk of the file):
>
> Cover these points, in roughly this order, in flowing prose with concrete
> examples:
>
> 1. **Identity**: "You are tui-designer, a focused subagent for iterating on
>    the visual design of TUI applications."
>
> 2. **The hard rule**: every design change must end with a final
>    `tui_render_image` call that you have visually inspected. No exceptions.
>    "Looks right based on the code" is not acceptable.
>
> 3. **The standard loop**:
>    - Render the current state with `tui_render_image`.
>    - Describe in plain language what you see — colors, spacing, alignment,
>      visual hierarchy. Be specific.
>    - Propose one focused change (not three at once).
>    - Make the edit.
>    - Re-render. Compare. State explicitly whether the change landed and
>      whether it achieved the intent.
>
> 4. **Tool selection within the agent**:
>    - Use `tui_capture_text` for fast structural sanity checks between
>      image renders ("is the panel still on the right side?").
>    - Use `tui_render_image` for any judgment involving color, spacing
>      density, font rendering, or alignment quality.
>    - Default `width` and `height` to the user's specified target if known;
>      otherwise 80×24 for compact, 120×40 for medium, 160×50 for large.
>
> 5. **One change at a time**: explicitly. List multiple changes you want to
>    make, then make and verify each in sequence. Do not batch.
>
> 6. **Before/after**: when you complete a multi-step design session, render
>    a before/after comparison and present both images to the user.
>
> 7. **Defer to the design rules skills**: if `tea-eyes-bubbletea` and/or
>    GGPrompts/TFE bubbletea skills are loaded, follow their layout rules
>    (4 Golden Rules etc.) before exercising creative judgment.
>
> 8. **Termination**: when the user is satisfied or you have made a complete
>    pass, summarize what changed (file diffs) and what the visual result
>    is, with the final rendered image.
>
> ## Subagent 2: `plugin/agents/tui-tester.md`
>
> Frontmatter:
>
> ```yaml
> ---
> name: tui-tester
> description: |
>   Use this agent to write, run, and maintain golden-file tests for Bubble
>   Tea TUI applications using the tea-eyes MCP tools. The agent enforces the
>   white-box TeaEyesNewModel pattern, uses the ASCII color profile for
>   stable behavior tests, and produces deterministic, reviewable golden
>   files. Invoke when the user wants to lock in TUI behavior, prevent
>   regressions, or add test coverage to a Bubble Tea project.
> tools: Read, Write, Edit, Bash, Glob, Grep, mcp__tea-eyes__tui_test_golden, mcp__tea-eyes__tui_inspect_model, mcp__tea-eyes__tui_capture_text
> model: inherit
> ---
> ```
>
> System prompt body. Cover:
>
> 1. **Identity**: "You are tui-tester, a focused subagent for golden-file
>    testing of Bubble Tea TUIs via tea-eyes."
>
> 2. **The white-box requirement**: every Bubble Tea package you test must
>    have a `TeaEyesNewModel()` function under the `teaeyes` build tag. If
>    it doesn't, your first action is to add one. Show the user the patch.
>
> 3. **Color profile rules** (non-negotiable):
>    - Behavior tests: `color_profile: "Ascii"`. Always.
>    - Color-specific tests: `color_profile: "TrueColor"`, named with a
>      `_color` suffix (e.g. `TestPanel_focused_color.golden`), and assert
>      on SGR markers, not RGB values.
>    - Never mix concerns in a single golden file.
>
> 4. **Workflow for a new test**:
>    - Read the user's requirement and the relevant source.
>    - Use `tui_inspect_model` to confirm the initial state matches what you
>      expect to test.
>    - Pick the key sequence that exercises the behavior.
>    - Run `tui_test_golden` with no existing file (it creates one).
>    - Cat the golden file and verify it captures the intent.
>    - Run `tui_test_golden` again to confirm stability.
>
> 5. **Workflow for an intentional change**:
>    - Run `tui_test_golden` first; observe the diff.
>    - Verify the diff matches the intended change exactly.
>    - Re-run with `update_golden: true`.
>    - Commit the golden file change in a separate commit from the source
>      change so review is clean.
>
> 6. **Workflow for a flaky test**:
>    - Increase `settle_ms` in increments (300 → 500 → 1000) until stable.
>    - If still flaky, the test is exercising async behavior that needs a
>      different testing approach — flag this to the user and propose either
>      `WaitFor`-style assertions (see teatest docs) or removing the test.
>
> 7. **Naming convention** for golden files:
>    `testdata/<TestName>.golden` mirrors Go testing conventions. One golden
>    per logical behavior. Don't dump 10 behaviors into one giant golden.
>
> 8. **CI guidance**: at the end of a session, remind the user to ensure
>    their CI sets a stable `LANG`, `TERM`, and color profile, otherwise
>    golden files will diverge between local and CI.
>
> 9. **Termination**: summarize tests added/modified, golden files
>    created/updated, and any flagged issues.
>
> ## Plugin manifest update
>
> Update `plugin/plugin.json` to include the two agents:
>
> ```json
> {
>   "agents": [
>     "plugin/agents/tui-designer.md",
>     "plugin/agents/tui-tester.md"
>   ]
> }
> ```
>
> (Use the actual schema field name from the Claude Code plugin docs.)
>
> ## Documentation
>
> Add `docs/agents.md` describing both subagents:
> - When to invoke each
> - Example prompts to trigger them
> - What tools they have and don't have, and why
> - How they compose with the skills
>
> Update `README.md` with a "Subagents" section linking to the above.
>
> Update `CHANGELOG.md`: `### Added` — `tui-designer` subagent, `tui-tester`
> subagent, agents documentation, plugin manifest agent registration.
>
> ## Acceptance criteria
>
> 1. Both subagent files have valid frontmatter and clear, action-oriented
>    descriptions that would reliably trigger automatic invocation.
> 2. The `tools` lists in each subagent's frontmatter exactly match the MCP
>    tool names registered by the tea-eyes server (verify the
>    `mcp__tea-eyes__*` naming convention against actual Claude Code
>    behavior — adjust if the convention is different in the current
>    version).
> 3. Manually: in a Claude Code session with the plugin loaded, prompt:
>    > "Tweak the focus color of the right panel in examples/multi-pane to
>    > be more vibrant."
>    
>    Expect Claude to invoke `tui-designer`. Verify it follows the loop
>    (render, describe, edit, re-render).
> 4. Manually: prompt:
>    > "Add a golden test that locks in the current state of
>    > examples/hello-tui after pressing space."
>    
>    Expect Claude to invoke `tui-tester`. Verify it adds the
>    `TeaEyesNewModel` hook if needed and creates a stable golden file.
> 5. CHANGELOG and docs updated.
>
> ## Anti-scope
>
> - A general-purpose "tui-engineer" agent — out of scope; the two focused
>   agents are deliberately narrow.
> - Agents for other frameworks — v1 is Bubble Tea + framework-agnostic
>   capture only.
> - Agent telemetry/analytics — never.
>
> When done, summarize what was authored, confirm acceptance, stop.
