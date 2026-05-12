# Phase 3 — Bubble Tea Plugin (teatest Integration)

**Goal:** Fast, in-process introspection for Bubble Tea apps specifically. This is the framework-aware plugin layer — agnostic core + BT plugin, as decided in planning.

**Estimated effort:** 1–2 days

**Prerequisites:** Phases 0–2 complete.

---

## Prompt to give Claude Code

> Implement Phase 3 of tea-eyes per `docs/plan.md`. Add a Bubble Tea-aware
> plugin layer using `github.com/charmbracelet/x/exp/teatest`. This is in-process
> rendering and message injection — orders of magnitude faster than pty or VHS,
> at the cost of being Bubble Tea-only and requiring a small white-box hook
> from the user's app.
>
> ## Dependencies
>
> Add:
> - `github.com/charmbracelet/x/exp/teatest`
> - `github.com/muesli/termenv` — for color profile control
>
> ## Design decisions
>
> Use the **white-box pattern**: the user adds a small `_teaeyes.go` file to
> their package that exports a `TeaEyesNewModel() tea.Model` function under a
> build tag. tea-eyes generates a tiny test harness binary that imports the
> user's package, calls that function, drives it with teatest, and dumps the
> result.
>
> This is awkward but it's the only way to inspect arbitrary Bubble Tea apps
> from the outside without forcing them to expose a debug port. Document the
> pattern thoroughly.
>
> Force ASCII color profile by default in golden output for determinism (per
> Carlos Becker's teatest guidance). Provide an opt-out for users who want to
> assert on color codes.
>
> ## Implementation
>
> ### 1. `internal/teatest/harness.go`
>
> Generate a temporary Go file that imports the user's package and calls their
> exported model constructor:
>
> ```go
> // generated harness — do not edit
> //go:build teaeyes
>
> package main
>
> import (
>     "encoding/json"
>     "io"
>     "os"
>
>     tea "github.com/charmbracelet/bubbletea"
>     "github.com/charmbracelet/x/exp/teatest"
>     "github.com/muesli/termenv"
>     "github.com/charmbracelet/lipgloss"
>
>     userpkg "{{ .ImportPath }}"
> )
>
> func main() {
>     lipgloss.SetColorProfile(termenv.{{ .ColorProfile }})
>
>     m := userpkg.{{ .ModelFunc }}()
>     tm := teatest.NewTestModel(nil, m,
>         teatest.WithInitialTermSize({{ .Width }}, {{ .Height }}),
>     )
>
>     // ... send key sequence (generated from inputs)
>
>     out, _ := io.ReadAll(tm.FinalOutput(nil))
>     result := map[string]any{
>         "final_output": string(out),
>         // optionally: model snapshot via reflection
>     }
>     _ = json.NewEncoder(os.Stdout).Encode(result)
> }
> ```
>
> Important details:
> - The `nil` passed to `NewTestModel` and `FinalOutput` is wrong because they
>   need a `*testing.T`. Work around by writing a minimal `testing.TB`
>   adapter, or invoke teatest from a `go test -run` harness — pick the
>   approach that works with current teatest API and document it.
> - Build the harness with `go build -tags teaeyes -o /tmp/<hash>` and execute
>   it. Cache the binary by hash of (import path + model func + Go file
>   contents).
>
> ### 2. `internal/teatest/driver.go`
>
> ```go
> type Driver struct { /* ... */ }
>
> type GoldenOpts struct {
>     PackagePath  string   // absolute or module-relative path to the user's package
>     ModelFunc    string   // exported function name, default "TeaEyesNewModel"
>     Keys         []string
>     Width        int
>     Height       int
>     ColorProfile string   // "Ascii", "ANSI", "ANSI256", "TrueColor"; default "Ascii"
>     GoldenFile   string   // path; if exists, compare; if not, create
>     UpdateGolden bool     // write current output to GoldenFile
> }
>
> type GoldenResult struct {
>     FinalOutput string
>     Match       bool
>     Diff        string  // unified diff if mismatch
> }
>
> func (d *Driver) RunGolden(ctx context.Context, opts GoldenOpts) (GoldenResult, error)
> ```
>
> Implementation:
> 1. Compute cache key from opts.
> 2. If harness binary not cached: generate harness, write to a temp dir,
>    build it.
> 3. Execute the harness with the key sequence as args (or env), capture
>    stdout, parse JSON result.
> 4. If `GoldenFile` doesn't exist or `UpdateGolden`: write final_output to
>    the file. Return Match=true.
> 5. Else: read GoldenFile, compare. If different, compute unified diff
>    (use `github.com/sergi/go-diff` or roll your own line diff).
>
> ### 3. `internal/teatest/inspect.go`
>
> A second tool that dumps the model state (best-effort) after N messages.
> Use Go reflection to JSON-encode the exported fields of the model struct.
>
> ```go
> type InspectOpts struct {
>     PackagePath string
>     ModelFunc   string
>     Keys        []string
>     Width       int
>     Height      int
> }
>
> type InspectResult struct {
>     ModelJSON string  // JSON of exported fields
>     ViewText  string  // current view as plain text
> }
>
> func (d *Driver) Inspect(ctx context.Context, opts InspectOpts) (InspectResult, error)
> ```
>
> Document clearly that only exported fields are visible. For unexported
> debugging, users should add their own debug methods.
>
> ### 4. MCP tools
>
> **Tool: `tui_test_golden`**
>
> Inputs:
> | name | type | required | default | description |
> |------|------|----------|---------|-------------|
> | `package_path` | string | yes | — | path to the Bubble Tea package |
> | `model_func` | string | no | "TeaEyesNewModel" | exported constructor |
> | `keys` | []string | no | `[]` | key sequence |
> | `width` | int | no | 80 | columns |
> | `height` | int | no | 24 | rows |
> | `color_profile` | enum | no | "Ascii" | Ascii/ANSI/ANSI256/TrueColor |
> | `golden_file` | string | yes | — | path to golden file |
> | `update_golden` | bool | no | false | overwrite golden |
>
> Output:
> ```json
> {
>   "match": true,
>   "diff": null,
>   "final_output_preview": "first 500 chars..."
> }
> ```
>
> **Tool: `tui_inspect_model`**
>
> Inputs: package_path, model_func, keys, width, height.
> Output: `{ "model_json": "...", "view_text": "..." }`
>
> ### 5. Update the example
>
> Add `examples/hello-tui/teaeyes.go`:
>
> ```go
> //go:build teaeyes
>
> package main
>
> import tea "github.com/charmbracelet/bubbletea"
>
> // TeaEyesNewModel is the entry point used by tea-eyes for in-process testing.
> // Build with `-tags teaeyes` to include.
> func TeaEyesNewModel() tea.Model {
>     return initialModel()
> }
> ```
>
> Add `examples/hello-tui/teaeyes_test.go` showing how a user would write
> a golden test by hand using the same teatest patterns (so users can choose
> either tea-eyes-driven golden tests or hand-written `go test` golden
> tests — the former is for Claude's iteration loop, the latter for CI).
>
> ### 6. `test/integration/teatest_test.go`
>
> Integration test:
> 1. Run `tui_test_golden` against `examples/hello-tui` with no golden file.
>    Assert match=true (creates the golden file).
> 2. Run again. Assert match=true (compares against existing).
> 3. Modify the harness inputs (different key sequence). Assert match=false
>    with a non-empty diff.
> 4. Run `tui_inspect_model`. Assert the returned `model_json` contains the
>    expected exported field names.
>
> ## Documentation updates
>
> - `docs/mcp-tools.md`: add `tui_test_golden` and `tui_inspect_model`
>   sections matching the format of the others.
> - `docs/workflow.md`: add a section "Locking in behavior with golden tests"
>   — when to use the in-process path vs. capture/render. Cover the speed/
>   coupling tradeoff.
> - New file `docs/white-box-pattern.md`: thorough explanation of the
>   `TeaEyesNewModel` convention, why the build tag, how to handle apps
>   with constructor arguments, how to handle non-`tea.Model` returns.
> - `CHANGELOG.md`: `### Added` — `tui_test_golden`, `tui_inspect_model`,
>   teatest harness generation, white-box pattern documentation.
>
> ## Acceptance criteria
>
> 1. All previous tests still pass; new integration tests pass.
> 2. Golden file workflow: delete the golden, run, assert created; run again,
>    assert matches; mutate inputs, assert clean diff produced.
> 3. Manually: in a Claude Code session, ask Claude to add a golden test for
>    a behavior in `examples/hello-tui`. Claude should use `tui_test_golden`
>    to bootstrap the golden file, then verify it stays stable.
> 4. The harness binary is cached and reused across calls with identical
>    package + model func.
> 5. CI green.
>
> ## Anti-scope
>
> - Skills/subagents — Phase 4/5
> - tmux driver — Phase 6
> - Anything beyond Bubble Tea (no Ratatui/Textual/Ink hooks)
> - Mutation testing or property-based testing
>
> When done, summarize, confirm acceptance, stop.
