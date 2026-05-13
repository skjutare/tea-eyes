# Compatibility with the GGPrompts / TFE bubbletea skill

[GGPrompts/TFE](https://github.com/GGPrompts/TFE) ships a
`bubbletea` Claude Code skill that encodes hard-won Bubble Tea layout
wisdom — the **4 Golden Rules**: account for borders, never auto-wrap in
bordered panels, match mouse detection to layout, use weights not pixels.

tea-eyes does **not** duplicate those rules. The two skills are
complementary.

| Question | Skill that answers it |
|---|---|
| "What should this layout look like?" | GGPrompts/TFE bubbletea |
| "Is the layout actually rendering that way?" | tea-eyes |
| "How do I know I haven't regressed?" | tea-eyes (`tui_test_golden`) |
| "What padding should I use here?" | GGPrompts/TFE bubbletea |
| "Did my padding change actually shift this panel?" | tea-eyes (`tui_capture_text`) |

## Install both

Order doesn't matter; both are skills and both load eagerly when relevant.

1. Install the GGPrompts/TFE bubbletea skill — clone the repo into your
   Claude Code skills directory or copy `.claude/skills/bubbletea/`
   from `github.com/GGPrompts/TFE` into your project.
2. Install the tea-eyes plugin — see the top-level `README.md`.
3. Open a Bubble Tea project in Claude Code. When you ask Claude to
   change anything visual, both skills should pull in: GGPrompts to
   guide the design decision, tea-eyes to verify the result.

## Quickstart with both

> I want to widen the left sidebar from 20 to 30 columns and ensure the
> right panel re-flows correctly.

What Claude does, with both skills loaded:

1. **GGPrompts** reminds: account for borders (so the inner usable
   width changes), check mouse-region calculations, use weights if the
   total width is dynamic.
2. **tea-eyes-loop** picks the right tool: `tui_capture_text` at the
   target size for a fast structural check.
3. Claude edits the constant.
4. **tea-eyes-loop** re-captures to verify the border landed where
   expected and the right panel didn't get clipped.
5. **tea-eyes-bubbletea** locks the change in with `tui_test_golden` at
   `color_profile: "Ascii"`.

Without GGPrompts, step 1 doesn't happen and the change might
miscalculate border accounting. Without tea-eyes, steps 2 + 4 + 5 don't
happen and a regression ships silently. Use both.

## Why this matters

The Bubble Tea layout footguns GGPrompts captures are subtle and easy to
re-introduce. tea-eyes catches the symptom (a panel one column off) but
GGPrompts prevents the cause. tea-eyes is not a substitute for design
expertise; it is a verifier for it.
