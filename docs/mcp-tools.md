# MCP tool reference

Each tool is documented here as it lands.

| Tool | Status | Lands in |
|------|--------|----------|
| `tui_capture_text` | available | Phase 1 |
| `tui_render_image` | available | Phase 2 |
| `tui_test_golden` | available | Phase 3 |
| `tui_inspect_model` | available | Phase 3 |
| `tui_session_attach_hint` | not yet | Phase 6 |

---

## `tui_capture_text`

Spawn a TUI command under a pseudo-terminal of fixed size, optionally send a
sequence of keystrokes, and return the rendered text grid. The pty driver
auto-replies to common terminal capability queries (CSI DSR, OSC 10/11, DA)
so libraries like termenv don't block on startup.

### Inputs

| name | type | required | default | description |
|------|------|----------|---------|-------------|
| `command` | string | yes | — | Binary path or command on `PATH` to run. |
| `args` | string[] | no | `[]` | Arguments to pass to the command. |
| `keys` | string[] | no | `[]` | Sequence of key specs to send after the initial render. |
| `width` | int | no | `80` | Terminal width in columns. |
| `height` | int | no | `24` | Terminal height in rows. |
| `settle_ms` | int | no | `300` | Milliseconds to wait after spawn and between key sends. |
| `strip_ansi` | bool | no | `true` | If `true`, return a clean text grid; if `false`, return raw bytes with SGR escapes intact. |
| `cwd` | string | no | `""` | Working directory for the spawned command. |

Key specs follow jmlago's tmux skill notation:

- Literal runes/strings: `"a"`, `"7"`, `"hello world"`
- Special keys: `"enter"`, `"escape"`, `"space"`, `"tab"`, `"backspace"`,
  `"up"`, `"down"`, `"left"`, `"right"`, `"home"`, `"end"`, `"pgup"`,
  `"pgdown"`, `"delete"`, `"insert"`, `"f1"`–`"f12"`
- Modifiers: `"ctrl+c"`, `"alt+x"`, `"ctrl+alt+l"`, `"shift+tab"`

### Output

Structured content:

```json
{
  "text": "...",
  "width": 80,
  "height": 24,
  "raw_bytes": 1234
}
```

`raw_bytes` is the size of the unprocessed pty output (useful for sanity
checks). Errors come back as MCP tool errors with actionable messages such as
`command not found`, `keys[2]="ctrl+meow": ...`, or
`process "foo" exited before initial settle`.

### Examples

Capture an idle screen of a Bubble Tea binary:

```json
{
  "command": "./examples/hello-tui",
  "width": 40,
  "height": 8
}
```

Send a keystroke and capture the result:

```json
{
  "command": "./examples/hello-tui",
  "keys": ["space"],
  "width": 40,
  "height": 8
}
```

Quit a TUI cleanly (the tool returns the rendered text up to the quit):

```json
{
  "command": "./examples/hello-tui",
  "keys": ["q"]
}
```

---

## `tui_render_image`

Render a TUI command as a PNG or GIF using [Charm's VHS](https://github.com/charmbracelet/vhs)
and return the image to the model as an MCP image content block. Use this
when text capture isn't enough — to judge colour, spacing, borders, focus
rings, or typography.

External dependencies: `vhs`, `ttyd`, `ffmpeg`. Run `tea-eyes doctor` to
verify they're on `PATH`.

### Inputs

| name | type | required | default | description |
|------|------|----------|---------|-------------|
| `command` | string | yes | — | Binary path or command on `PATH` to run. |
| `args` | string[] | no | `[]` | Arguments to pass to the command. |
| `keys` | string[] | no | `[]` | Key sequence to send after spawn (same notation as `tui_capture_text`). |
| `width` | int | no | `80` | Terminal width in columns. |
| `height` | int | no | `24` | Terminal height in rows. |
| `font_family` | string | no | `"JetBrains Mono"` | Monospace font name installed locally. |
| `font_size` | int | no | `14` | Font size in points. |
| `theme` | string | no | `"Dracula"` | VHS theme (see `vhs themes`). |
| `format` | enum | no | `"png"` | `"png"` (still frame) or `"gif"` (whole session). |
| `padding` | int | no | `20` | Pixel padding around the terminal. |
| `settle_ms` | int | no | `300` | Milliseconds to wait between key sends. |
| `no_cache` | bool | no | `false` | Bypass the on-disk render cache. |
| `cwd` | string | no | `""` | Working directory for the spawned command. |

### Caching

Renders are cached under `$XDG_CACHE_HOME/tea-eyes/renders/` (or
`~/.cache/tea-eyes/renders/`) keyed by a SHA-256 of the canonicalized
inputs. Each entry is the image plus a sibling `.tape` file for debugging.
The cache is opportunistic — there is no staleness check; run
`tea-eyes cache clean` to clear it.

### Output

An MCP image content block plus structured metadata:

```json
{
  "format": "png",
  "width": 80,
  "height": 24,
  "mime": "image/png",
  "bytes": 18432,
  "cache_hit": false,
  "cache_path": "/Users/me/.cache/tea-eyes/renders/abc123.png"
}
```

### Examples

Render the first frame of a Bubble Tea binary:

```json
{
  "command": "./examples/multi-pane",
  "width": 80,
  "height": 24
}
```

Animate a key sequence as a GIF:

```json
{
  "command": "./examples/multi-pane",
  "keys": ["tab", "tab", "q"],
  "format": "gif",
  "settle_ms": 400
}
```

Render with a custom font and theme:

```json
{
  "command": "./examples/hello-tui",
  "keys": ["space"],
  "font_family": "Fira Code",
  "font_size": 16,
  "theme": "GitHub Dark"
}
```

---

## `tui_test_golden`

Run a Bubble Tea program in-process via the [teatest](https://github.com/charmbracelet/x/tree/main/exp/teatest)
harness and compare its final terminal output against a golden file. This is
the fast feedback loop: ~10ms per run, no pty, no VHS, but Bubble Tea-only
and requires a small white-box hook in the user's package — see
[white-box-pattern.md](./white-box-pattern.md).

### Inputs

| name | type | required | default | description |
|------|------|----------|---------|-------------|
| `package_path` | string | yes | — | Filesystem path to the Bubble Tea package directory. |
| `model_func` | string | no | `"TeaEyesNewModel"` | Exported constructor that returns a `tea.Model`. |
| `keys` | string[] | no | `[]` | Key sequence sent before capturing final output. |
| `width` | int | no | `80` | Terminal columns. |
| `height` | int | no | `24` | Terminal rows. |
| `color_profile` | enum | no | `"Ascii"` | One of `Ascii`, `ANSI`, `ANSI256`, `TrueColor`. Ascii is deterministic. |
| `golden_file` | string | yes | — | Path to the golden file. Created automatically on first run. |
| `update_golden` | bool | no | `false` | If `true`, overwrite the golden with the current output. |

### Output

```json
{
  "match": true,
  "created": true,
  "diff": "",
  "golden_path": "/path/to/file.golden",
  "final_output_preview": "first 500 chars…"
}
```

When `match` is `false`, `diff` contains a small unified-style diff between
the golden and the captured output. The harness binary is cached under
`$XDG_CACHE_HOME/tea-eyes/teatest/` keyed by package contents and
`model_func`, so subsequent calls skip the `go test -c` step.

### Examples

Bootstrap a golden file for the example app:

```json
{
  "package_path": "./examples/hello-tui",
  "keys": ["space"],
  "width": 40,
  "height": 8,
  "golden_file": "./examples/hello-tui/golden/space.txt"
}
```

Regenerate after an intentional UI change:

```json
{
  "package_path": "./examples/hello-tui",
  "keys": ["space"],
  "golden_file": "./examples/hello-tui/golden/space.txt",
  "update_golden": true
}
```

---

## `tui_inspect_model`

Drive a Bubble Tea program in-process and return the JSON-encoded **exported
fields** of the final model plus the current `View()` output. Useful for
asserting on state transitions without having to parse the rendered text.

Unexported fields are not visible — add a debug accessor or temporarily
export the field you care about.

### Inputs

| name | type | required | default | description |
|------|------|----------|---------|-------------|
| `package_path` | string | yes | — | Path to the Bubble Tea package. |
| `model_func` | string | no | `"TeaEyesNewModel"` | Exported constructor. |
| `keys` | string[] | no | `[]` | Key sequence before snapshotting. |
| `width` | int | no | `80` | Terminal columns. |
| `height` | int | no | `24` | Terminal rows. |
| `color_profile` | string | no | `"Ascii"` | Color profile. |

### Output

```json
{
  "model_json": "{\"Counter\":1,\"ShowCounter\":true}",
  "view_text": "Hello, tea-eyes!\nCounter: 1\n",
  "final_output_preview": "…"
}
```
