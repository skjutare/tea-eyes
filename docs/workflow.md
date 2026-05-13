# Workflow

This document is fleshed out in Phase 7. Until then, here are the iterative
loops tea-eyes enables today.

## Iterating on visual design with image rendering

Once `tui_render_image` is wired up (Phase 2+), the design loop looks like:

1. **Render.** Ask Claude to run `tui_render_image` against the example
   binary you're iterating on.
2. **Look.** Claude receives a real PNG and can describe what it sees:
   misaligned borders, low-contrast text, awkward spacing.
3. **Reason.** Discuss what to change — colour, padding, layout — in plain
   language.
4. **Edit.** Claude edits the Bubble Tea / Lipgloss code.
5. **Re-render.** Re-run `tui_render_image`. Pass `no_cache: true` if you
   changed inputs outside the captured options (fonts, themes set globally,
   etc.); otherwise the cache makes re-renders free for unchanged inputs.

### Performance tradeoff vs `tui_capture_text`

| Tool | Typical latency | Use when |
|------|-----------------|----------|
| `tui_capture_text` | ~50 ms | iterating on logic, structure, or grid-level layout |
| `tui_render_image` (cached) | ~5 ms | re-checking an unchanged screen |
| `tui_render_image` (cold) | 2–5 s | judging pixels — colour, focus rings, typography |

A good loop alternates: drive the TUI with text capture while the
*structure* is wrong, then switch to image capture once the structure is
right and only the *aesthetics* remain.
