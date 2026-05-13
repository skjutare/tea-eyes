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

## Locking in behavior with golden tests

Once a TUI behaves correctly, lock it in with `tui_test_golden`. Compared
to the capture/render path:

| Path | Speed | Fidelity | Coupling |
|------|-------|----------|----------|
| `tui_capture_text` (pty) | ~50 ms | text grid | none — runs your binary |
| `tui_render_image` (VHS) | 2–5 s cold | pixels | external binaries (vhs/ttyd/ffmpeg) |
| `tui_test_golden` (teatest) | ~10 ms | text grid | Bubble Tea + `teaeyes` build tag |

Use the in-process path when you want a regression-suite-style assertion:
"this key sequence, against this model, should always produce this output."
The white-box hook (`TeaEyesNewModel`, see
[white-box-pattern.md](./white-box-pattern.md)) gives the harness a way to
construct your model without launching the real binary.

A typical loop:

1. Iterate on the TUI using `tui_capture_text` and/or `tui_render_image`.
2. Once the screen looks right, ask Claude to call `tui_test_golden` with
   no existing golden — the first call creates the file.
3. From then on, the same call protects against regressions. When you
   *intentionally* change the UI, re-run with `update_golden: true`.

`tui_inspect_model` is the complement: rather than asserting on the
rendered output, it dumps the exported fields of the final model so you can
make claims about state transitions ("after pressing tab three times,
`SelectedIndex` should be 2").
