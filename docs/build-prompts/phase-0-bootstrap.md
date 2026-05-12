# Phase 0 — Repo Bootstrap

**Goal:** Empty but well-formed repo on GitLab (mirrored to GitHub). Anyone can `git clone` and the structure tells them what's coming. No application code yet — just scaffolding, license, CI, docs skeleton.

**Estimated effort:** ~½ day

**Prerequisites:**
- Go 1.26+ installed
- `git` configured for both GitLab and GitHub
- A new empty repo created on `gitlab.com/<your-namespace>/tea-eyes`
- A new empty repo created on `github.com/<your-namespace>/tea-eyes`
- (Optional but recommended) GitLab → GitHub push mirror configured under
  Settings → Repository → Mirroring repositories

---

## Prompt to give Claude Code

> Bootstrap a new Go module for **tea-eyes** — a Claude Code plugin providing
> visual feedback for TUI development. The MCP server (Go) plus companion skills
> and reference subagents will let Claude Code "see" what it's building when
> developing terminal user interfaces, with first-class support for the Charm
> ecosystem (Bubble Tea, Lipgloss, Bubbles).
>
> **Do not write any application code in this phase.** This phase is scaffolding
> only. All source files should be placeholders (interfaces or empty package
> declarations) so that the layout is visible but no logic is implemented.
>
> ## Module + license
>
> - Go module path: `gitlab.com/<MY-NAMESPACE>/tea-eyes` (replace placeholder).
> - Go version: 1.26 (toolchain directive).
> - License: **MIT**, mirroring Charmbracelet's wording. Copyright line:
>   `Copyright (c) 2026 Christoffer Skjutare`
> - Add a `NOTICE` file crediting the prior art this project builds on:
>   - The Charm ecosystem (Bubble Tea, Lipgloss, VHS, teatest) — MIT
>   - GGPrompts/TFE bubbletea Claude Code skill
>   - jmlago "Debug TUIs with tmux" Claude Code skill
>   - rigerc/bubbletea-v2-scaffold
>   - Hatchet's blog post "Building a TUI is easy now"
>
> ## Directory layout
>
> Create exactly this structure with placeholder files (`.gitkeep` where empty,
> minimal `doc.go` for Go packages):
>
> ```
> tea-eyes/
> ├── LICENSE
> ├── NOTICE
> ├── README.md
> ├── CONTRIBUTING.md
> ├── CHANGELOG.md                         # "## [Unreleased]" only
> ├── Makefile
> ├── go.mod
> ├── .gitignore                           # Go-standard + .tape, *.gif, *.png in test outputs
> ├── .editorconfig
> ├── .gitlab-ci.yml
> ├── .github/
> │   ├── workflows/
> │   │   └── ci.yml                       # mirror of GitLab CI, simpler
> │   ├── ISSUE_TEMPLATE/
> │   │   ├── bug_report.md
> │   │   └── feature_request.md
> │   └── pull_request_template.md
> ├── .gitlab/
> │   └── issue_templates/
> │       ├── Bug.md
> │       └── Feature.md
> ├── cmd/
> │   └── tea-eyes/
> │       └── main.go                       # package main, empty main() with TODO
> ├── internal/
> │   ├── server/doc.go                     # MCP server wiring (mark3labs/mcp-go)
> │   ├── pty/doc.go                        # pty-based process driver (Phase 1)
> │   ├── tmux/doc.go                       # tmux-based driver (Phase 6)
> │   ├── capture/doc.go                    # text capture
> │   ├── render/doc.go                     # vhs wrapping for PNG/GIF (Phase 2)
> │   ├── teatest/doc.go                    # bubble tea plugin (Phase 3)
> │   └── keys/doc.go                       # key string parser
> ├── pkg/
> │   └── teaeyes/doc.go                    # public Go API (so others can embed)
> ├── plugin/
> │   ├── plugin.json                       # placeholder, populated in Phase 4
> │   ├── skills/
> │   │   ├── tea-eyes-loop/.gitkeep
> │   │   └── tea-eyes-bubbletea/.gitkeep
> │   └── agents/.gitkeep
> ├── examples/
> │   ├── hello-tui/.gitkeep                # Phase 1 will populate
> │   └── multi-pane/.gitkeep               # Phase 2 will populate
> ├── docs/
> │   ├── plan.md                           # COPY THE FULL MULTI-PHASE PLAN HERE
> │   ├── architecture.md                   # placeholder, populated in Phase 7
> │   ├── mcp-tools.md                      # placeholder, populated as tools land
> │   ├── workflow.md                       # placeholder, populated in Phase 7
> │   └── compat-ggprompts.md               # placeholder, populated in Phase 4
> └── test/
>     ├── integration/.gitkeep
>     └── golden/.gitkeep
> ```
>
> ## README.md content
>
> Sections, in order:
>
> 1. **Title + one-line tagline**: "Visual feedback for TUI development with
>    Claude Code."
> 2. **Status badge**: `alpha — under construction`.
> 3. **What this is**: a Claude Code plugin (MCP server + skills + subagents)
>    that lets Claude Code see what it's building when developing TUIs. Three
>    capture/render strategies: tmux text capture, VHS image rendering, teatest
>    in-process golden output.
> 4. **What this is not**: a replacement for the GGPrompts/TFE bubbletea skill
>    (which encodes layout best practices). tea-eyes is *additive* — install
>    both for the full design loop.
> 5. **Why**: links to the Hatchet "TUIs are easy now" blog post and a one-line
>    summary of the existing tmux-driven workflow, then "tea-eyes packages this
>    pattern as a first-class MCP server."
> 6. **Status / roadmap**: bulleted list of the 8 phases (0–7) with checkboxes,
>    all unchecked except Phase 0.
> 7. **Prior art and credits**: link list to Charm, GGPrompts/TFE, jmlago,
>    rigerc, Hatchet, mark3labs/mcp-go.
> 8. **License**: MIT. Reference NOTICE.
>
> Keep it scannable. No screenshots yet.
>
> ## CONTRIBUTING.md content
>
> - This is a side project; no SLA on issues or PRs.
> - Conventional Commits encouraged.
> - Run `make test lint` before pushing.
> - PRs that affect the MCP tool surface need a CHANGELOG entry under
>   `## [Unreleased]`.
> - Code of conduct: defer to the Contributor Covenant 2.1 (link, don't copy).
>
> ## Makefile targets
>
> ```
> build       # go build ./cmd/tea-eyes
> test        # go test ./...
> lint        # go vet ./... && staticcheck ./...
> tidy        # go mod tidy
> snapshot    # goreleaser release --snapshot --clean (requires goreleaser)
> clean       # rm -rf dist/ test/golden/*.actual
> ```
>
> ## .gitlab-ci.yml
>
> Stages: `lint`, `test`, `build`. Use `golang:1.26` image. Cache the Go module
> cache. Run `go vet`, `staticcheck`, `go test -race ./...`, then a goreleaser
> snapshot build on tag pushes only.
>
> ## .github/workflows/ci.yml
>
> Mirror of the above but simpler: a single `ci` job that runs `make lint test`.
> No release pipeline (GitLab is canonical for releases).
>
> ## docs/plan.md
>
> Copy the full multi-phase build plan into this file (I'll provide separately,
> or use the `docs/plan.md` I supplied with this prompt). This becomes the
> single source of truth for what's being built.
>
> ## go.mod
>
> Just the module declaration and Go version. No dependencies yet.
>
> ## Acceptance criteria (do not finish until these all pass)
>
> 1. `make build` succeeds (builds an empty binary).
> 2. `make test` succeeds (no tests yet, but the command runs without error).
> 3. `make lint` succeeds.
> 4. `git status` is clean after `git add . && git commit`.
> 5. The directory tree exactly matches the layout above.
> 6. Pushing to GitLab triggers the CI pipeline and it goes green.
> 7. The README links resolve (no 404s).
>
> ## Out of scope for Phase 0
>
> - Any MCP server logic
> - Any TUI capture or rendering
> - Any skills or subagent files (just the directory placeholders)
> - Any examples (just the directory placeholders)
>
> When you're done, summarize what you built and confirm all acceptance
> criteria pass. Then stop and wait — Phase 1 is a separate prompt.
