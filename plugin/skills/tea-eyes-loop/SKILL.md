---
name: tea-eyes-loop
description: |
  Use this skill when developing, modifying, or debugging any terminal user
  interface (TUI) application — Bubble Tea, Ratatui, Textual, Ink, or
  anything else that renders to a terminal. Teaches Claude how to use the
  tea-eyes MCP tools (tui_capture_text, tui_render_image, tui_test_golden,
  tui_inspect_model) to see what the TUI actually renders, judge layout and
  visual quality, and iterate. Trigger any time you would otherwise be
  guessing what the TUI looks like — e.g. when changing layout, colors,
  spacing, key bindings, or any visual element.
license: MIT
---

# tea-eyes loop

You are editing a TUI. You cannot see it. The tea-eyes MCP tools are how you
see it. Use them.

## The loop

```
capture  →  reason about what's on screen  →  edit code  →  re-capture
```

After **every** visual change, re-capture. Reading a diff is not enough — a
one-character style tweak can shift a whole panel. If you skip the
re-capture step you will ship a broken layout and not know it.

## Which tool when

| Goal | Tool | Why |
|------|------|-----|
| Quick structural check ("is the panel where I expect?") | `tui_capture_text` | ~50 ms. ASCII grid. Free. |
| Color, spacing, font, or typography judgment | `tui_render_image` | Real pixels. Only way to judge true color, border weight, padding. |
| Lock in a Bubble Tea behavior so it can't regress | `tui_test_golden` | Persistent assertion. Diffs on mismatch. |
| Inspect runtime model state during debugging | `tui_inspect_model` | Structured dump of exported model fields and `View()` output. |

Default to `tui_capture_text`. Reach for `tui_render_image` only when the
question is genuinely about pixels (colors, fonts, spacing, themes). Reach
for `tui_test_golden` / `tui_inspect_model` only for Bubble Tea apps that
expose `TeaEyesNewModel()` under the `teaeyes` build tag.

## Anti-patterns

- **Rendering an image to check layout.** `tui_render_image` takes seconds
  and burns cache; `tui_capture_text` answers most layout questions in tens
  of milliseconds. If you can't tell whether a panel is in the right place
  from ASCII, you have a layout bug, not a tool problem.
- **Iterating without re-capturing.** You edited the `View()` function;
  prove it. Don't claim the change works because the diff looks right.
- **Not pinning `width` and `height`.** Different sizes produce different
  renders. Always pass explicit values; never rely on defaults if the
  output will be compared between calls.
- **Too-short `settle_ms`.** TUIs with async commands (HTTP, timers,
  spinners) need time to reach their final frame. If a capture shows a
  loading state you didn't expect, raise `settle_ms` (1000–3000 ms).
- **Theme drift in golden files.** A behavior test that captures
  truecolor ANSI bytes will break the moment a theme changes. For behavior
  goldens, set `color_profile: "Ascii"`. Reserve `TrueColor` for tests
  that are explicitly about color.

## A standard loop (worked example)

You're widening the left sidebar in a Bubble Tea app from 20 to 30 columns.

1. **Baseline capture.**
   ```
   tui_capture_text({ command: "./myapp", width: 100, height: 30, settle_ms: 500 })
   ```
   Note the current border position around column 21.

2. **Edit.** Change the sidebar width constant.

3. **Re-capture** with the same `width`/`height`. Border is now at
   column 31. Body content shifts right. Confirm no horizontal overflow.

4. **Render an image** if the change involves color or padding decisions.
   Otherwise skip.

5. **Lock it in** if this layout matters for the long term:
   ```
   tui_test_golden({
     package_path: "./internal/ui",
     model_func: "TeaEyesNewModel",
     golden_file: "test/golden/sidebar-wide.txt",
     width: 100, height: 30,
     color_profile: "Ascii"
   })
   ```
   Commit `sidebar-wide.txt`. Future regressions will surface as a diff.

5 turns, no guesswork.

## Composing with other skills

`tea-eyes-loop` teaches the *feedback mechanism*. It does not encode design
rules.

- If `tea-eyes-bubbletea` is installed, follow it for Bubble Tea
  specifics (the white-box pattern, color-profile discipline).
- If the **GGPrompts/TFE bubbletea** skill is installed, follow its
  layout rules (the 4 Golden Rules: account for borders; never auto-wrap
  in bordered panels; match mouse detection to layout; use weights not
  pixels). tea-eyes does not duplicate those — it verifies them.

When a framework-specific skill disagrees with this one on workflow
sequence or tool choice, the framework-specific skill wins.
