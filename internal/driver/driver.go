// Package capturedrv defines the abstraction over capture backends — pty and
// tmux — so MCP tools can take a `mode` parameter and stay agnostic to which
// process model is in play.
package capturedrv

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitlab.com/skjutare/tea-eyes/internal/pty"
	"gitlab.com/skjutare/tea-eyes/internal/tmux"
)

// Mode identifies a capture backend.
type Mode string

const (
	// ModePTY spawns the TUI under a fresh pseudo-terminal owned by the
	// tea-eyes process. Fast, dependency-free, ephemeral.
	ModePTY Mode = "pty"
	// ModeTmux drives the TUI inside a tmux session that the user can attach
	// to in another terminal. Requires the tmux binary on PATH.
	ModeTmux Mode = "tmux"
)

// ParseMode returns the Mode for s, defaulting to ModePTY when s is empty.
func ParseMode(s string) (Mode, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "pty":
		return ModePTY, nil
	case "tmux":
		return ModeTmux, nil
	default:
		return "", fmt.Errorf("driver: unknown mode %q (want \"pty\" or \"tmux\")", s)
	}
}

// CaptureOpts is the union of inputs accepted by both backends. Fields that do
// not apply to the chosen mode are ignored.
type CaptureOpts struct {
	Command  string
	Args     []string
	Keys     [][]byte
	Width    int
	Height   int
	SettleMs int
	Cwd      string

	// SessionName is the tmux session name to use. Empty means create a fresh
	// ephemeral session. Ignored when Mode is pty.
	SessionName string
	// Persist asks tmux mode to leave the session alive after Capture returns.
	// Ignored when Mode is pty.
	Persist bool
}

// CaptureResult is what Capture returns. Raw is the wire bytes that downstream
// renderers (e.g. internal/capture) decode into a text grid.
type CaptureResult struct {
	Raw []byte
	// Session, when non-empty, is the tmux session name that survived the
	// call. Populated only for tmux mode with Persist=true (or when the
	// caller supplied an existing SessionName).
	Session string
}

// Driver captures TUI output from a backend.
type Driver interface {
	Capture(ctx context.Context, opts CaptureOpts) (CaptureResult, error)
}

// New returns a Driver for the requested mode. tmux mode is verified eagerly
// so the user gets a clear error before any TUI is spawned.
func New(mode Mode) (Driver, error) {
	switch mode {
	case ModePTY:
		return &ptyDriver{inner: pty.New()}, nil
	case ModeTmux:
		td := tmux.New()
		if _, err := td.LookPath(); err != nil {
			return nil, err
		}
		return &tmuxDriver{inner: td}, nil
	default:
		return nil, fmt.Errorf("driver: unsupported mode %q", mode)
	}
}

type ptyDriver struct{ inner *pty.Driver }

func (d *ptyDriver) Capture(ctx context.Context, opts CaptureOpts) (CaptureResult, error) {
	raw, err := d.inner.Capture(ctx, pty.SpawnOpts{
		Command:  opts.Command,
		Args:     opts.Args,
		Width:    opts.Width,
		Height:   opts.Height,
		SettleMs: opts.SettleMs,
		Cwd:      opts.Cwd,
	}, opts.Keys)
	return CaptureResult{Raw: raw}, err
}

type tmuxDriver struct{ inner *tmux.Driver }

const tmuxSpawnSettleMin = 250 * time.Millisecond

// Capture creates or reuses a tmux session, runs the user's command in its
// pane, waits for the initial render, delivers each key sequence (with a
// settle pause between sends), captures the pane, and tears down the session
// unless the caller asked to keep it alive.
//
//nolint:gocognit // intrinsic: session lifecycle + key loop + cleanup
func (d *tmuxDriver) Capture(ctx context.Context, opts CaptureOpts) (CaptureResult, error) {
	if opts.Command == "" {
		return CaptureResult{}, errors.New("tmux: command is required")
	}
	settle := max(time.Duration(opts.SettleMs)*time.Millisecond, tmuxSpawnSettleMin)

	// Build the command line tmux will execute directly (bypassing the user's
	// login shell so slow rc files don't delay the TUI).
	var sb strings.Builder
	sb.WriteString(opts.Command)
	for _, a := range opts.Args {
		sb.WriteString(" ")
		sb.WriteString(shellQuote(a))
	}
	cmdLine := sb.String()

	// Persistent sessions need to stay alive after the TUI exits so the user
	// (or a follow-up call) can attach. `remain-on-exit on` plus
	// `respawn-pane` on reuse gives us that semantics without spawning a
	// shell.
	keep := opts.Persist || opts.SessionName != ""

	var (
		session   *tmux.Session
		err       error
		reused    bool
		sessionEx bool
	)
	if opts.SessionName != "" {
		sessionEx, err = d.inner.HasSession(ctx, opts.SessionName)
		if err != nil {
			return CaptureResult{}, err
		}
	}
	switch {
	case opts.SessionName != "" && sessionEx:
		session, err = d.inner.AttachExisting(ctx, opts.SessionName)
		reused = true
	default:
		session, err = d.inner.NewSession(ctx, tmux.SessionOpts{
			Name:         opts.SessionName,
			Width:        opts.Width,
			Height:       opts.Height,
			Cwd:          opts.Cwd,
			Command:      cmdLine,
			RemainOnExit: keep,
		})
	}
	if err != nil {
		return CaptureResult{}, err
	}

	defer func() {
		if !keep {
			_ = session.Close(context.Background())
		}
	}()

	if reused {
		if rcErr := session.RespawnPane(ctx, cmdLine); rcErr != nil {
			return CaptureResult{Session: keepName(session, keep)}, rcErr
		}
	}
	if waitErr := sleepCtx(ctx, settle); waitErr != nil {
		return CaptureResult{Session: keepName(session, keep)}, waitErr
	}

	for i, k := range opts.Keys {
		if len(k) == 0 {
			continue
		}
		if skErr := session.SendKeys(ctx, k); skErr != nil {
			return CaptureResult{Session: keepName(session, keep)}, fmt.Errorf("tmux: key #%d: %w", i, skErr)
		}
		if waitErr := sleepCtx(ctx, settle); waitErr != nil {
			return CaptureResult{Session: keepName(session, keep)}, waitErr
		}
	}

	// Always capture with colors so downstream code can choose to strip them.
	text, capErr := session.CapturePane(ctx, true)
	if capErr != nil {
		return CaptureResult{Session: keepName(session, keep)}, capErr
	}
	return CaptureResult{Raw: []byte(text), Session: keepName(session, keep)}, nil
}

func keepName(s *tmux.Session, keep bool) string {
	if keep {
		return s.Name
	}
	return ""
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// shellQuote wraps s in single quotes, escaping any embedded single quotes.
// Good enough for the simple binary+args case; not a general-purpose shell
// escape.
func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	if isSafeShellWord(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func isSafeShellWord(s string) bool {
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '_', r == '-', r == '/', r == '.', r == '=', r == ':':
		default:
			return false
		}
	}
	return true
}
