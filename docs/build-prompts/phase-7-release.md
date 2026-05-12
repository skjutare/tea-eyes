# Phase 7 — Polish, Docs, Release v0.1.0

**Goal:** Ship-ready v0.1.0. Long-form tutorial with screenshots, architecture docs, polished README, cross-platform release artifacts, and submissions to the major Claude Code skill/plugin directories.

**Estimated effort:** 1–2 days

**Prerequisites:** Phases 0–6 complete. All tests green. Working plugin
locally.

---

## Prompt to give Claude Code

> Implement Phase 7 of tea-eyes per `docs/plan.md`. This is the polish-and-
> release phase. No new functional code; everything is documentation,
> packaging, and distribution.
>
> ## 1. Long-form tutorial: `docs/workflow.md`
>
> Rewrite as a complete narrative. The reader follows along start to finish:
>
> 1. **Setup** (5 min): install tea-eyes, install vhs/ttyd/ffmpeg/tmux,
>    register the plugin with Claude Code, verify with `tea-eyes doctor`.
> 2. **Bootstrap a Bubble Tea app** (10 min): scaffold a minimal app with two
>    panels.
> 3. **First design pass with the loop** (15 min): use the `tui-designer`
>    subagent to refine the layout. Show the actual Claude Code transcript
>    excerpts where the agent renders, describes, edits, re-renders.
>    Embed PNGs at each step (record them with vhs from the actual session).
> 4. **Locking it in with a golden test** (10 min): use the `tui-tester`
>    subagent to add the `TeaEyesNewModel` hook and create golden files for
>    the layout and key behaviors.
> 5. **Watching the agent** (5 min): demonstrate `mode=tmux` +
>    `tmux_persist=true`, with a screenshot of two terminals side by side.
> 6. **Composing with GGPrompts** (5 min): install the GGPrompts/TFE
>    bubbletea skill alongside, repeat one of the design steps, show how the
>    rules from GGPrompts steered the agent toward correct layout.
> 7. **What's next**: pointer to `docs/architecture.md` for extension,
>    `CONTRIBUTING.md` for contribution.
>
> Embed images. Generate them with vhs (eat your own dog food). Commit them
> under `docs/img/`.
>
> ## 2. Architecture doc: `docs/architecture.md`
>
> - Block diagram (use Mermaid for git-renderability):
>   - Claude Code → MCP stdio → tea-eyes server
>   - server → driver (pty | tmux) → user TUI process
>   - server → render (vhs subprocess) → PNG/GIF
>   - server → teatest harness → user package (compiled with build tag)
> - Per-package responsibilities (table: package → responsibility → key
>   types/funcs).
> - Extension points: how to add a new driver, how to add a new render
>   backend, how to add a framework plugin (e.g. for Ratatui in v2).
> - Error handling philosophy: every tool error must be actionable.
> - Caching strategy: what's cached, what's keyed on what, how to invalidate.
> - Concurrency model: each MCP call is independent; no shared state across
>   tool calls except cache and tmux sessions.
>
> ## 3. MCP tools reference: `docs/mcp-tools.md`
>
> Final pass. For each of the five tools (`tui_capture_text`,
> `tui_render_image`, `tui_test_golden`, `tui_inspect_model`,
> `tui_session_attach_hint`):
>
> - One-paragraph summary
> - Full input table with types, defaults, descriptions
> - Full output schema
> - Three example invocations (simple, intermediate, complex)
> - Common errors and what they mean
> - Performance notes (typical latency, when caching helps)
>
> ## 4. README polish
>
> Final pass. Sections in this order:
>
> 1. **Logo / banner** (optional ASCII art using charm-style aesthetic).
> 2. **One-line tagline + status**: now `beta` instead of `alpha`.
> 3. **Demo GIF** at the top: 10-second screencap of the design loop in
>    action, generated with vhs.
> 4. **What it is** (3 bullet sentences).
> 5. **Install** (one section per platform: macOS via brew tap, Linux via
>    `go install`, manual build). Include `claude mcp add` invocation and
>    `claude plugin install` invocation.
> 6. **Quickstart** (the 5-minute happy path: render a hello world TUI).
> 7. **Subagents** with example prompts for each.
> 8. **Skills** with composition note re: GGPrompts.
> 9. **MCP tools** (one-line summary of each, link to docs).
> 10. **Status / roadmap** with all 8 phases checked.
> 11. **Prior art and credits** (full list, links).
> 12. **License** (MIT).
>
> Keep total length under ~400 lines. Move depth to dedicated docs.
>
> ## 5. Release engineering
>
> - Add `.goreleaser.yaml` configured for:
>   - Builds: linux/amd64, linux/arm64, darwin/amd64, darwin/arm64,
>     windows/amd64
>   - Archive: `tar.gz` for unix, `zip` for windows; include LICENSE,
>     NOTICE, README, plugin/ directory
>   - Checksums: SHA-256
>   - Changelog: from `CHANGELOG.md` `## [Unreleased]` section
>   - Homebrew tap formula (optional; configure to push to a separate
>     `homebrew-tea-eyes` repo)
> - Update `.gitlab-ci.yml`: tag pushes trigger `goreleaser release` (not
>   snapshot). Use a CI variable for the GitHub token (mirrored release).
> - Update `.github/workflows/release.yml` to also run goreleaser on tag,
>   pushing to GitHub Releases. Both releases publish in parallel from
>   their respective remotes.
> - Set `VERSION` in `cmd/tea-eyes/main.go` (or via ldflags) so
>   `tea-eyes --version` reports the build version.
>
> Cut the v0.1.0 release:
> 1. Move `## [Unreleased]` content to `## [0.1.0] - YYYY-MM-DD` in
>    CHANGELOG.
> 2. Tag `v0.1.0`, push to GitLab.
> 3. Verify both GitLab and GitHub release artifacts.
>
> ## 6. Distribution submissions
>
> Draft submission text (place in `docs/submissions/`) for each directory:
>
> - **mcpmarket.com** — submit as MCP server with skill bundle.
> - **claude-plugins.dev** — submit as Claude Code plugin (preferred listing
>   path).
> - **fastmcp.me** — submit as agent skill collection.
> - **smithery.ai** — submit as MCP server.
> - **mcp.directory** — submit if their schema allows.
>
> Each submission text includes: tagline, longer description, install
> command, link to repo, link to demo GIF, license, author. Save as
> separate files so the user can copy-paste each.
>
> ## 7. Launch announcement (draft)
>
> Write `docs/announcement-draft.md` — a blog-post-style announcement
> mirroring the framing of Hatchet's "Building a TUI is easy now" post.
> Sections:
>
> - Hook: "What if Claude Code could see your TUI?"
> - The problem (TUI development feedback loops are slow without visual
>   tooling).
> - The solution (visual feedback loop via MCP, modeled on Playwright for
>   browsers).
> - How it composes with the existing ecosystem (Bubble Tea, GGPrompts skill,
>   jmlago tmux skill, rigerc scaffold) — explicit credit.
> - The standard loop, with an embedded GIF.
> - "Try it" — install command.
> - "What's next" — roadmap.
> - Credits and thanks.
>
> Don't post anywhere yet. Just have it ready.
>
> ## 8. Final QA pass
>
> Before tagging, do a manual QA pass:
>
> 1. Fresh checkout in a clean directory.
> 2. `make build test lint` — green.
> 3. `make test-integration` (with vhs+tmux installed) — green.
> 4. Install the plugin into Claude Code via the documented install command.
> 5. Run through `docs/workflow.md` step by step. Every command works,
>    every screenshot is current.
> 6. `tea-eyes doctor` — all dependencies reported correctly.
> 7. Each subagent triggers reliably from natural-language prompts.
> 8. Each MCP tool returns sensible output and useful errors when called
>    incorrectly.
>
> Fix anything that breaks. Re-tag if needed.
>
> ## Acceptance criteria
>
> 1. v0.1.0 tagged and released on both GitLab and GitHub with artifacts.
> 2. README has a working demo GIF at the top.
> 3. `docs/workflow.md` is a complete narrative tutorial with embedded
>    images.
> 4. `docs/architecture.md` has a Mermaid block diagram and per-package
>    table.
> 5. `docs/mcp-tools.md` is a complete reference for all five tools.
> 6. Submission drafts exist for mcpmarket, claude-plugins, fastmcp,
>    smithery, mcp.directory.
> 7. Launch announcement draft exists.
> 8. Manual QA pass complete; all steps documented in workflow.md verified.
> 9. CHANGELOG `[0.1.0]` entry complete and accurate.
>
> ## Anti-scope
>
> - Submitting to directories — that's a manual step the user does after
>   reviewing the drafts.
> - Posting the announcement — manual.
> - v0.2 features — keep them in `docs/roadmap.md` only.
>
> When done, list everything shipped, give the exact commands the user
> should run to (a) verify the release locally and (b) submit to each
> directory, and stop.
