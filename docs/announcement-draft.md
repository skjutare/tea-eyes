# Launch announcement (draft)

> Draft. Not posted yet. Tone modeled on Hatchet's "Building a TUI is
> easy now" post — concrete, credit-heavy, ends with a demo.

---

## What if Claude Code could see your TUI?

When Claude Code builds a web app, it can use Playwright — render the
page, see the DOM, click buttons, screenshot the result, iterate. When
Claude Code builds a TUI, it has historically been blind. It writes
Bubble Tea code, hopes the layout works, and you have to run the app and
report back.

That feedback loop is slow. Worse, it's noisy: "the right panel is too
wide, no the left one, no the *spacing* is off" — five round trips later
you're still describing pixels in English.

**tea-eyes** closes that loop. It's a Claude Code plugin — an MCP
server, two skills, and two subagents — that lets Claude *see* what
it's building.

## How

Three complementary capture/render strategies, each tuned for a
different job:

| Strategy | What it gives you | When to use it |
|---|---|---|
| **Text capture** via pty (or tmux) | Fast ASCII grid (~50 ms) | Iterating on structure, layout, key handling |
| **Image render** via VHS | True-pixel PNG/GIF (2–5 s cold, ~5 ms cached) | Judging color, spacing density, borders, focus rings |
| **In-process introspection** via teatest | Structured grid + JSON state (~10 ms) | Bubble Tea apps; locking behaviors in with golden files |

Claude picks the right one per task. The packaged skill teaches when to
pick which; the two reference subagents bake the choice into their tool
list.

## The standard loop

```
render → look → reason → edit → re-render
```

The `tui-designer` subagent enforces a hard rule: every design change
ends with a final `tui_render_image` call that the agent has visually
inspected. "Looks right based on the code I just wrote" is not accepted.
Lipgloss styles compose in non-obvious ways and terminal cell widths
bite — pixels are the only proof.

The `tui-tester` subagent enforces a different discipline: the white-box
`TeaEyesNewModel` pattern (so the harness drives the *model*, not the
*binary*), and strict color-profile rules (ASCII for behavior tests,
TrueColor for color-specific tests, never mixed).

## How it composes with the existing ecosystem

tea-eyes is **additive**. It does not replace any of the following — it
extends them.

- **[Bubble Tea, Lipgloss, VHS, teatest](https://charm.sh)** by Charm —
  the foundation. tea-eyes is a thin orchestration layer; all the heavy
  lifting belongs to Charm. License matches (MIT), credits are explicit.
- **GGPrompts/TFE bubbletea skill** — encodes the "4 Golden Rules" of
  Bubble Tea layout. The `tea-eyes-bubbletea` skill **defers** all
  layout guidance to it. Install both for the full design loop.
- **[jmlago "Debug TUIs with tmux"](https://github.com/jmlago)** skill —
  the ad-hoc tmux pattern that validated the workflow. tea-eyes
  formalizes the same pattern as a typed MCP tool surface; both coexist.
- **rigerc/bubbletea-v2-scaffold** — project-shaped scaffold; tea-eyes
  is plugin-shaped. They compose — start from the scaffold, install
  tea-eyes as a Claude Code plugin.
- **[Hatchet's "Building a TUI is easy now"](https://hatchet.run)** post
  — validated the approach (Claude + tmux + Bubble Tea is genuinely
  productive). This announcement explicitly credits and links it.

## Try it

```sh
# install (Go 1.26+)
go install gitlab.com/skjutare/tea-eyes/cmd/tea-eyes@latest

# install the external dependencies (macOS)
brew install vhs ttyd ffmpeg tmux

# verify
tea-eyes doctor

# register with Claude Code as a full plugin
claude plugin install gitlab.com/skjutare/tea-eyes
```

Then ask Claude something like:

> *Tweak the focus color of the right panel in `./examples/multi-pane` to
> be more vibrant. Use the tui-designer subagent.*

…and watch the loop happen.

[Full walkthrough →](https://gitlab.com/skjutare/tea-eyes/-/blob/main/docs/workflow.md)

## What's next

- v0.2 — framework plugins (Ratatui first; Textual and Ink to follow)
- v0.3 — visual diffing with bounding boxes
- v0.4 — record-and-replay sessions as regression tests
- v1.0 — API stability after ≥3 months of v0.x in active use

Full roadmap: [`docs/roadmap.md`](https://gitlab.com/skjutare/tea-eyes/-/blob/main/docs/roadmap.md).

## Credits

The whole point of tea-eyes is to be additive to a stack that's already
working. Explicit thanks to Charm, GGPrompts/TFE, jmlago, rigerc, the
mark3labs/mcp-go maintainers, and the Hatchet team for the post that
made this feel inevitable.

— Christoffer Skjutare
