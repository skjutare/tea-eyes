# Workflow

A start-to-finish walkthrough. By the end you'll have:

- tea-eyes installed and wired into Claude Code,
- a two-panel Bubble Tea app,
- a `tui-designer` session that refines its layout visually,
- a `tui-tester` session that locks the behavior in with a golden file,
- and a tmux-attached terminal where you can watch Claude work live.

Estimated time: ~45 minutes the first run, ~10 minutes once you've done it.

If you get stuck, every command in this tutorial is verified against a
fresh checkout in the Phase 7 QA pass. Open an issue if something diverges.

---

## 1. Setup (~5 min)

### Install tea-eyes

Pick one:

```sh
# from source (any platform)
go install gitlab.com/skjutare/tea-eyes/cmd/tea-eyes@latest

# or download a release artifact
# (macOS / Linux / Windows; amd64 + arm64) from the Releases page
```

Verify:

```sh
tea-eyes -version
```

### Install the external dependencies

`tui_render_image` shells out to [`vhs`](https://github.com/charmbracelet/vhs),
which itself needs `ttyd` and `ffmpeg`. `tmux` is optional and only needed
if you want to watch Claude drive the TUI live.

```sh
brew install vhs ttyd ffmpeg tmux
# Linux / Windows: see https://github.com/charmbracelet/vhs#installation
```

Run the doctor — it tells you exactly what's missing and how to fix it:

```sh
tea-eyes doctor
```

Output should end with `0 required dependency(ies) missing`.

### Register with Claude Code

As an MCP server (minimum):

```sh
claude mcp add tea-eyes -- tea-eyes serve
```

As a full plugin (recommended — gets you the skills and subagents too):

```sh
claude plugin install gitlab.com/skjutare/tea-eyes
```

Open Claude Code in any project and ask "what tea-eyes tools do you have?"
— you should see `tui_capture_text`, `tui_render_image`, `tui_test_golden`,
`tui_inspect_model`, and `tui_session_attach_hint`.

---

## 2. Bootstrap a Bubble Tea app (~10 min)

If you already have a Bubble Tea app, skip ahead. Otherwise we'll use
[`examples/multi-pane`](../examples/multi-pane) — a minimal two-panel demo
that exercises the layout primitives most TUIs need.

```sh
git clone https://gitlab.com/skjutare/tea-eyes
cd tea-eyes
go build -o ./bin/multi-pane ./examples/multi-pane
```

Sanity check that the binary runs:

```sh
./bin/multi-pane     # press q to quit
```

You should see two side-by-side panels and a status bar.

---

## 3. First design pass with the loop (~15 min)

This is the headline workflow: render, look, reason, edit, re-render.

Open Claude Code in the `tea-eyes` directory and prompt:

> *Tweak the right panel of `./bin/multi-pane` — the focus border feels
> washed out, and the title above it is too quiet. Use the tui-designer
> subagent.*

What you should see:

1. **Claude invokes `tui-designer`.** The subagent's `description` field is
   tuned to trigger on "tweak / design / focus / border / spacing" phrasing.
2. **It renders the current state** via `tui_render_image` at a sensible
   size and **describes it in plain language** — which panel has focus, what
   color the border is, how the title sits above.
3. **It proposes one focused change** (not three) and edits the Bubble Tea
   / Lipgloss source.
4. **It re-renders and compares**, telling you explicitly whether the
   change landed. If a `Lipgloss` style didn't apply cleanly, it iterates
   on the edit — it doesn't paper over with a second change.

When the agent is satisfied, it presents a final before/after pair. The
"before" image is the very first render it took; the "after" is the last.
You can see the delta without reading a single diff.

![multi-pane example](img/multi-pane-demo.png)

> **Why image, not text?** Lipgloss composes styles in non-obvious ways and
> terminal cell widths bite. The only proof that a design change landed is
> the rendered pixels. The agent enforces this hard rule: it never
> declares a design change done without a final `tui_render_image` it has
> visually inspected. See [`agents.md`](./agents.md#tui-designer) for the
> full contract.

### Cache hits

The second render of an unchanged screen is ~5 ms — `tui_render_image`
hashes the canonicalized inputs and serves from
`$XDG_CACHE_HOME/tea-eyes/renders/`. Clear it with `tea-eyes cache clean`
when you change something outside the captured options (system fonts,
global theme).

---

## 4. Lock it in with a golden test (~10 min)

Now that the layout is right, prevent regressions. Same Claude Code
session, new prompt:

> *Add a golden test that locks in the current state of
> `./examples/multi-pane` after pressing tab once (which moves focus to
> the right panel). Use the tui-tester subagent.*

What happens:

1. **Claude invokes `tui-tester`.**
2. **It checks for `TeaEyesNewModel`.** If the package doesn't have the
   white-box hook yet, it adds one under the `teaeyes` build tag and
   shows you the patch:

   ```go
   //go:build teaeyes

   package multipane

   import tea "github.com/charmbracelet/bubbletea"

   func TeaEyesNewModel() tea.Model {
       return newModel( /* deterministic args */ )
   }
   ```

   The hook constructs the model with deterministic inputs — no
   `time.Now()`, no randomness — otherwise goldens would be flaky. See
   [`white-box-pattern.md`](./white-box-pattern.md) for the full pattern.

3. **It picks the shortest key sequence** that exercises the behavior
   (`["tab"]`), and calls `tui_test_golden` with no existing file. The
   tool creates one and returns `match: true, created: true`.

4. **It `cat`s the golden** and confirms it captures the intent —
   nothing surprising, no stray timestamps, no extra frames.

5. **It runs `tui_test_golden` again** to confirm stability. Two
   consecutive passes = trustworthy golden.

The agent uses `color_profile: "Ascii"` by default for behavior tests —
that's non-negotiable and decouples the golden from theme changes. If you
also want to lock in colors, ask explicitly: *"add a color-specific test
for the focused panel using TrueColor"* — the agent will create a separate
golden with the `_color` suffix and assert on SGR markers, not raw RGB
values.

### Updating goldens after an intentional change

When you deliberately change the UI:

> *I changed the focus border style. Update the golden to match.*

The agent runs the test, observes the diff, verifies the diff matches the
intended change exactly, then re-runs with `update_golden: true`. It
recommends committing the golden change in a separate commit from the
source change so reviewers can see the diff cleanly.

---

## 5. Watching the agent (~5 min)

For trust-building, debugging, or a demo-style workflow you can run the
TUI inside a tmux session and attach to it in another terminal.

In Claude Code:

> *Run `./bin/multi-pane` in a persistent tmux session, give me the
> attach command, then send tab three times.*

Claude's flow:

1. `tui_capture_text` with `mode: "tmux"`, `tmux_persist: true`, sensible
   size. The result includes `tmux_session: "teaeyes-1f9c"`.
2. `tui_session_attach_hint` with that name. Returns the exact shell
   command: `tmux attach -t teaeyes-1f9c`.
3. Subsequent `tui_capture_text` calls with the same `tmux_session` name
   reuse the session (`respawn-pane`).

In your second terminal:

```sh
tmux attach -t teaeyes-1f9c
```

The TUI is right there, in its last state. Watch each step Claude takes.
Detach with `Ctrl-b d`. Kill with `tmux kill-session -t teaeyes-1f9c`.

Caveats:

- tmux mode is text-only — `tui_render_image` rejects `mode="tmux"`
  because VHS records its own pty internally.
- The session runs the binary directly, bypassing your login shell, so
  slow rc files don't delay startup.

---

## 6. Composing with the GGPrompts/TFE bubbletea skill (~5 min)

tea-eyes is *additive*. The GGPrompts/TFE bubbletea Claude Code skill
encodes the "4 Golden Rules" of Bubble Tea layout (account for borders,
never auto-wrap in bordered panels, match mouse detection to layout, use
weights not pixels). tea-eyes verifies via real renders; GGPrompts
prescribes what good looks like.

Install both and the `tui-designer` agent defers to GGPrompts for layout
rules before exercising creative judgment. If a design choice fights the
rules, the agent surfaces the conflict instead of silently deviating.

See [`compat-ggprompts.md`](./compat-ggprompts.md) for the full
composition story.

---

## What's next

- [`architecture.md`](./architecture.md) — package layout, extension
  points, design contracts. Read if you want to add a driver, a render
  backend, or a framework plugin.
- [`mcp-tools.md`](./mcp-tools.md) — full reference for every MCP tool,
  every input, every output, every error mode.
- [`agents.md`](./agents.md) — when to invoke each subagent, example
  prompts, the tool lists (including what they deliberately don't have).
- [`roadmap.md`](./roadmap.md) — what's coming after v0.1.0.
- [`CONTRIBUTING.md`](../CONTRIBUTING.md) — best-effort SLA, how to file
  good issues.
