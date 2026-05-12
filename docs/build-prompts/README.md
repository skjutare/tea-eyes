# tea-eyes — Build Prompts

This directory contains the multi-phase build plan for **tea-eyes**, a Claude
Code plugin providing visual feedback for TUI development with first-class
support for the Charm ecosystem.

**Author:** Christoffer Skjutare
**License:** MIT (matching Charm)
**Working name:** tea-eyes (alt: tasseo, scry, leafread)

---

## How to use these prompts

Each phase is a self-contained prompt to give to Claude Code on your dev
machine. Run them in order. Each phase ends in a working, demoable state — you
can stop after any phase and still have something useful.

### Suggested workflow per phase

1. Open a fresh Claude Code session in the `tea-eyes/` repo.
2. Paste the entire phase prompt.
3. Let Claude work; intervene only when stuck.
4. Verify the acceptance criteria yourself.
5. Commit, push, tag intermediate milestones if you want.
6. Move to the next phase in a new session (clean context).

### Tips

- Keep `docs/plan.md` (the master plan) in the repo as a reference Claude
  can re-read whenever it needs orientation.
- Each prompt explicitly lists "anti-scope" items — features explicitly
  deferred to later phases. If Claude tries to over-build, point at the
  anti-scope.
- Acceptance criteria are non-negotiable; don't move on until they all
  pass.

---

## The phases

| # | File | Goal | Effort |
|---|------|------|--------|
| 0 | `phase-0-bootstrap.md` | Repo skeleton, license, CI, docs placeholders | ½ day |
| 1 | `phase-1-mcp-pty.md` | MCP server + pty driver + `tui_capture_text` | 1–2 days |
| 2 | `phase-2-vhs-render.md` | VHS image rendering + `tui_render_image` | 1–2 days |
| 3 | `phase-3-teatest.md` | Bubble Tea plugin + `tui_test_golden` + `tui_inspect_model` | 1–2 days |
| 4 | `phase-4-skills.md` | Two SKILL.md files + plugin manifest + GGPrompts compat | ½–1 day |
| 5 | `phase-5-subagents.md` | `tui-designer` + `tui-tester` reference subagents | ½ day |
| 6 | `phase-6-tmux.md` | tmux driver + `mode` parameter on tools | 1 day |
| 7 | `phase-7-release.md` | Polish, docs, v0.1.0 release | 1–2 days |

**Total: ~7–10 working days**, comfortable for a part-time effort over a few
weekends.

---

## Architecture summary

After all phases:

```
Claude Code
   │ (stdio MCP)
   ▼
tea-eyes server (Go, mark3labs/mcp-go)
   ├── tui_capture_text  ──► driver (pty | tmux) ──► user TUI
   ├── tui_render_image  ──► vhs subprocess       ──► PNG/GIF
   ├── tui_test_golden   ──► teatest harness       ──► user package
   ├── tui_inspect_model ──► teatest harness       ──► user package
   └── tui_session_attach_hint
   
plugin/
   ├── skills/tea-eyes-loop          (framework-agnostic loop guidance)
   ├── skills/tea-eyes-bubbletea     (additive to GGPrompts/TFE skill)
   └── agents/{tui-designer, tui-tester}
```

---

## Composition with the existing ecosystem

tea-eyes is **additive**, not a replacement:

- **GGPrompts/TFE bubbletea skill** — encodes layout rules (4 Golden Rules,
  borders, mouse handling). tea-eyes defers to it for design rules.
- **jmlago "Debug TUIs with tmux" skill** — the original tmux-driven workflow.
  tea-eyes formalizes it as an MCP server.
- **rigerc/bubbletea-v2-scaffold** — project scaffold. tea-eyes plays nicely
  alongside.
- **Charm's Bubble Tea, Lipgloss, VHS, teatest** — the foundation. tea-eyes
  is a thin orchestration layer; all the heavy lifting is theirs.

Credit them all in NOTICE, README, and the launch announcement.

---

## Open questions to revisit before Phase 0

- Final repo name (tea-eyes vs alternatives)?
- GitLab namespace?
- Mirror to GitHub or single source?
- Brew tap or Go install only?

These don't block starting — Phase 0 uses placeholders that are easy to
search-and-replace.
