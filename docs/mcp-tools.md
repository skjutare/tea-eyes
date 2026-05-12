# MCP tool reference

Each tool is documented here as it lands.

| Tool | Status | Lands in |
|------|--------|----------|
| `tui_capture_text` | available | Phase 1 |
| `tui_render_image` | not yet | Phase 2 |
| `tui_test_golden` | not yet | Phase 3 |
| `tui_inspect_model` | not yet | Phase 3 |
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
