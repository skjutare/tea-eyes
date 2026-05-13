# The white-box pattern

`tui_test_golden` and `tui_inspect_model` need to construct an instance of
your Bubble Tea `tea.Model` from outside the process. Bubble Tea programs
are normal Go binaries, so there is no debug port to attach to — instead,
tea-eyes asks you to publish a tiny **white-box hook** in your package.

This page explains the convention, why it exists, and how to handle the
awkward cases.

## The hook

Add one file to your Bubble Tea package. The minimum:

```go
//go:build teaeyes

package yourpkg

import tea "github.com/charmbracelet/bubbletea"

// TeaEyesNewModel is the entry point tea-eyes uses for in-process testing.
// Only compiled when the `teaeyes` build tag is set, so it never affects
// production builds.
func TeaEyesNewModel() tea.Model {
    return initialModel()
}
```

Two things are load-bearing:

1. **The `//go:build teaeyes` tag.** Without it, this function would compile
   into your production binary. With it, normal `go build` and `go test`
   skip the file entirely.
2. **The function signature `func() tea.Model`.** tea-eyes calls it with no
   arguments and expects a value satisfying `tea.Model`.

The name is configurable per-call (`model_func`), but `TeaEyesNewModel` is
the conventional default.

## Why a build tag?

Without the tag, `TeaEyesNewModel` would be a public API surface that ships
to users. Tagging it `teaeyes` keeps it out of every other compilation: the
function only exists when tea-eyes asks for it.

## What tea-eyes does with it

1. Generates a `_test.go` file inside your package containing a single test
   function that imports teatest, sets up a `tea.Program` driven by your
   `TeaEyesNewModel()`, and prints the captured output as JSON.
2. Compiles the package with `go test -tags teaeyes -c` and caches the
   resulting binary under `$XDG_CACHE_HOME/tea-eyes/teatest/`.
3. Removes the generated file. Your source tree is left untouched between
   calls.
4. Invokes the cached binary, passing keys / size / color profile via
   environment variables. The same binary is reused for every subsequent
   call with the same package.

The cache key is a SHA-256 of `(package_path, model_func, .go file contents,
go.mod, go.sum)` — change any of those and the next call rebuilds.

## Handling constructor arguments

If your real model takes config (theme, port, DB handle, …) write a helper
that fills in defaults for testing:

```go
//go:build teaeyes

package yourpkg

import tea "github.com/charmbracelet/bubbletea"

func TeaEyesNewModel() tea.Model {
    return New(Config{
        Width:   80,
        Height:  24,
        DataDir: "testdata",
    })
}
```

If you want multiple test fixtures, expose multiple functions
(`TeaEyesEmptyState`, `TeaEyesPopulatedState`, …) and call them by passing
`model_func` to the MCP tool.

## What if my constructor doesn't return `tea.Model`?

It must, for the harness to compile. The usual workaround is a thin
adapter:

```go
//go:build teaeyes

package yourpkg

import tea "github.com/charmbracelet/bubbletea"

func TeaEyesNewModel() tea.Model {
    m, _ := initialModelWithError()  // unwrap or ignore the secondary return
    return m
}
```

## Color profile

For deterministic golden output the harness forces `lipgloss.SetColorProfile`
to `termenv.Ascii` by default. Use `color_profile: "TrueColor"` if you
specifically want to assert on SGR escape codes — but be aware that exact
RGB values across versions of lipgloss/bubbletea may drift, so prefer to
assert on the *presence* of color markers rather than precise sequences.

## Inspecting state

`tui_inspect_model` returns a JSON object containing only the **exported
fields** of the final model. Unexported fields are invisible by design — if
your test needs to assert on internal state, either:

- Export the field temporarily (the build tag keeps it out of production).
- Add a `Debug() map[string]any` method on your model and have
  `TeaEyesNewModel` return a wrapper that exposes it.

## What about `package main`?

This works fine. Bubble Tea apps written as `package main` (like
`examples/hello-tui`) can still define `TeaEyesNewModel`; the generated
test file lives in the same package and references the model directly.
