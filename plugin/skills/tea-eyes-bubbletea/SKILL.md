---
name: tea-eyes-bubbletea
description: |
  Use this skill when building or modifying Bubble Tea TUI applications in Go
  with the tea-eyes MCP tools available. Adds Bubble Tea-specific guidance to
  the tea-eyes-loop skill: the white-box TeaEyesNewModel pattern for
  in-process testing, color profile guidance for stable golden files, the
  teatest workflow, and image rendering for visual review. This skill is
  ADDITIVE to the GGPrompts/TFE bubbletea skill (which encodes layout rules)
  — if both are installed, defer all layout/component design rules to that
  skill and use this one only for the tea-eyes-specific workflow.
license: MIT
---

# tea-eyes for Bubble Tea

> **Composition.** This skill is **additive** to the GGPrompts/TFE
> bubbletea skill. If that skill is installed, follow its **4 Golden
> Rules** for layout (account for borders; never auto-wrap in bordered
> panels; match mouse detection to layout; use weights not pixels). This
> skill does **not** duplicate or override those rules. It only adds the
> tea-eyes workflow on top.

## 1. The white-box pattern

To use `tui_test_golden` and `tui_inspect_model`, expose your model
constructor to tea-eyes via a build-tagged file.

Create `teaeyes.go` in the package that holds your model:

```go
//go:build teaeyes

package myapp

import tea "github.com/charmbracelet/bubbletea"

// TeaEyesNewModel returns a model in its default initial state for
// in-process testing under the `teaeyes` build tag.
func TeaEyesNewModel() tea.Model {
    return NewModel() // or model{} — whatever your constructor is
}
```

The `//go:build teaeyes` tag keeps this file out of normal `go build` /
`go test`. tea-eyes injects a one-shot test driver alongside it, compiles
with `-tags teaeyes`, runs the test, then removes the harness.

If your real constructor takes arguments, expose a default-construct
wrapper here. The point of `TeaEyesNewModel` is to give tea-eyes a
deterministic starting state.

## 2. Color profile discipline

`color_profile` is the single most important knob for stable goldens.

| Test kind | `color_profile` | Why |
|---|---|---|
| **Behavior** — layout, key bindings, content | `Ascii` | Strips all styling. Goldens survive theme changes. |
| **Color** — verifying a status got highlighted, a header is bold | `TrueColor` | Emits real SGR escapes. Assert on patterns (e.g. `\x1b[48;2;`), never on specific RGB. |
| Anything else | `Ascii` | When in doubt, prefer stability. |

**Never** mix behavior and color assertions in one golden file. They have
opposite stability requirements — keep them in separate goldens.

## 3. The teatest workflow

First time creating a golden:

```jsonc
tui_test_golden({
  package_path: "./internal/ui",
  model_func:   "TeaEyesNewModel",
  golden_file:  "test/golden/main-view.txt",
  keys:         ["tab", "down", "down", "enter"],
  width: 80, height: 24,
  color_profile: "Ascii"
})
// → match=true, created=true. Commit the golden file.
```

Subsequent runs:

```jsonc
tui_test_golden({ ...same args, no update_golden })
// → match=true on no change; match=false with unified diff on regression.
```

Intentional change:

```jsonc
tui_test_golden({ ...same args, update_golden: true })
// → rewrites the file. Review the git diff before committing.
```

`tui_inspect_model` complements goldens — use it while debugging to dump
the exported model fields and the final `View()` text without writing
anything to disk.

## 4. Image rendering for design

When the question is "does it *look* right" rather than "is it
correct," use `tui_render_image`. The renderer is VHS-backed, so themes
are first-class:

```jsonc
tui_render_image({ command: "./myapp", theme: "Dracula" })
tui_render_image({ command: "./myapp", theme: "Solarized Dark" })
tui_render_image({ command: "./myapp", theme: "GruvboxDark" })
```

Compare side by side to pick a default theme or verify a contrast change.
Renders are cached on disk keyed by the full input set; repeat calls are
free.

## 5. Common Bubble Tea gotchas tea-eyes catches

- **Initial render diverges from later renders.** Capture immediately
  after spawn vs. after a no-op key. If they differ, your `Init()` is
  returning a cmd whose result you depend on for the first paint.
- **`tea.WindowSizeMsg` not handled.** Capture at 80×24 and at 120×40 —
  if the layout doesn't respond, you're not subscribing to size
  changes.
- **Async cmds need longer `settle_ms`.** If a capture shows a spinner
  or "loading…" you didn't expect, bump `settle_ms` (1000–3000 ms) or
  send a key that triggers the awaited transition.
- **Reflection only sees exported fields.** `tui_inspect_model` shows
  fields with capitalized names. Lowercase model state is invisible —
  if you need to debug it, export it or add a getter.

## 6. See also

For all **layout, border, mouse, and component design rules** — how to
structure panels, where to put padding, when to use weights vs. fixed
sizes — see the **GGPrompts/TFE bubbletea skill**. tea-eyes verifies the
rendered output; GGPrompts teaches you what the rendered output should
look like in the first place.
