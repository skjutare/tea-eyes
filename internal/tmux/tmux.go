package tmux

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Driver wraps a tmux binary. The zero value is not usable; construct via New.
type Driver struct {
	tmuxBin string
}

// New returns a Driver that uses the tmux binary on PATH. The lookup is
// deferred until first use so that the doctor command can introspect
// availability without forcing every server start to depend on tmux.
func New() *Driver { return &Driver{tmuxBin: "tmux"} }

// NewWithBinary lets tests substitute a specific tmux path.
func NewWithBinary(path string) *Driver { return &Driver{tmuxBin: path} }

// SessionOpts describes how to create a new tmux session.
type SessionOpts struct {
	Name   string // session name; if empty a "teaeyes-<rand>" name is generated
	Width  int    // initial pane width; defaults to 80
	Height int    // initial pane height; defaults to 24
	Cwd    string // working directory for the initial shell; defaults to caller's
	// Command, if set, is run directly as the session's initial process
	// instead of the user's login shell. This dodges slow rc files (oh-my-zsh
	// etc.) so the TUI starts immediately.
	Command string
	// RemainOnExit, if true, keeps the pane open after the initial Command
	// exits so callers can send further keys or attach for inspection.
	RemainOnExit bool
}

// Session is a handle to a tmux session created by or attached to via this
// driver. Owned sessions are killed by Close; existing sessions are left alone.
type Session struct {
	Name    string
	Owned   bool // true if we created it (and may kill it on cleanup)
	tmuxBin string
}

const (
	defaultWidth  = 80
	defaultHeight = 24
	// sendKeysFixedArgs is the number of fixed argv entries before the hex
	// payload list in `tmux send-keys -t <name> -H ...`.
	sendKeysFixedArgs = 4
)

// LookPath returns the resolved path of the tmux binary, or an error with an
// actionable install hint. Used by the doctor subcommand.
func (d *Driver) LookPath() (string, error) {
	bin := d.tmuxBin
	if bin == "" {
		bin = "tmux"
	}
	path, err := exec.LookPath(bin)
	if err != nil {
		return "", fmt.Errorf(
			"tmux: binary %q not found on PATH; install with `brew install tmux` "+
				"(or your package manager's equivalent). tmux is optional — pty "+
				"mode does not require it",
			bin,
		)
	}
	return path, nil
}

// HasSession reports whether a tmux session with the given name exists.
func (d *Driver) HasSession(ctx context.Context, name string) (bool, error) {
	if name == "" {
		return false, errors.New("tmux: session name is required")
	}
	//nolint:gosec // G204: tmux args are constructed from validated inputs
	out, err := exec.CommandContext(ctx, d.tmuxBin, "has-session", "-t", name).CombinedOutput()
	if err == nil {
		return true, nil
	}
	// tmux exits 1 when the session does not exist; treat that as "no" rather
	// than an error.
	if ee := new(exec.ExitError); errors.As(err, &ee) {
		return false, nil
	}
	return false, fmt.Errorf("tmux has-session: %w (%s)", err, strings.TrimSpace(string(out)))
}

// NewSession creates a fresh detached session of the requested size. If
// opts.Name is empty a random "teaeyes-<hex>" name is generated.
func (d *Driver) NewSession(ctx context.Context, opts SessionOpts) (*Session, error) {
	if _, err := d.LookPath(); err != nil {
		return nil, err
	}
	if opts.Width <= 0 {
		opts.Width = defaultWidth
	}
	if opts.Height <= 0 {
		opts.Height = defaultHeight
	}
	name := opts.Name
	if name == "" {
		var err error
		name, err = randomSessionName()
		if err != nil {
			return nil, err
		}
	} else if !validSessionName(name) {
		return nil, fmt.Errorf("tmux: invalid session name %q (allowed: letters, digits, '_', '-')", name)
	}

	args := []string{
		"new-session", "-d",
		"-s", name,
		"-x", strconv.Itoa(opts.Width),
		"-y", strconv.Itoa(opts.Height),
	}
	if opts.Cwd != "" {
		args = append(args, "-c", opts.Cwd)
	}
	if opts.Command != "" {
		args = append(args, opts.Command)
	}
	//nolint:gosec // G204: tmux args are validated above
	out, err := exec.CommandContext(ctx, d.tmuxBin, args...).CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf(
			"tmux new-session -s %q: %w (%s); is tmux installed and on PATH?",
			name, err, strings.TrimSpace(string(out)),
		)
	}
	if opts.RemainOnExit {
		//nolint:gosec // G204: validated session name
		ro, roErr := exec.CommandContext(ctx, d.tmuxBin,
			"set-option", "-t", name, "remain-on-exit", "on",
		).CombinedOutput()
		if roErr != nil {
			//nolint:gosec // G204: validated session name; rollback after failed set-option
			_ = exec.CommandContext(ctx, d.tmuxBin, "kill-session", "-t", name).Run()
			return nil, fmt.Errorf(
				"tmux set-option remain-on-exit: %w (%s)",
				roErr, strings.TrimSpace(string(ro)),
			)
		}
	}
	return &Session{Name: name, Owned: true, tmuxBin: d.tmuxBin}, nil
}

// AttachExisting returns a Session handle for an existing tmux session,
// without taking ownership. Close on the returned handle is a no-op.
func (d *Driver) AttachExisting(ctx context.Context, name string) (*Session, error) {
	if _, err := d.LookPath(); err != nil {
		return nil, err
	}
	ok, err := d.HasSession(ctx, name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf(
			"tmux: session %q does not exist. Use `tmux ls` to list sessions, or "+
				"omit tmux_session to create a fresh one",
			name,
		)
	}
	return &Session{Name: name, Owned: false, tmuxBin: d.tmuxBin}, nil
}

// RunCommand sends the given shell command followed by Enter to the session's
// active pane. Use this to spawn the TUI under test.
func (s *Session) RunCommand(ctx context.Context, cmd string) error {
	if cmd == "" {
		return errors.New("tmux: RunCommand requires a non-empty command")
	}
	//nolint:gosec // G204: arguments are explicit subcommands, not shell-evaluated
	out, err := exec.CommandContext(ctx, s.tmuxBin,
		"send-keys", "-t", s.Name, cmd, "Enter",
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux send-keys (run): %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// RespawnPane replaces whatever is running in the session's active pane with
// cmd, killing the previous process if any. Used to drive a fresh command in
// a session created on a prior call.
func (s *Session) RespawnPane(ctx context.Context, cmd string) error {
	if cmd == "" {
		return errors.New("tmux: RespawnPane requires a non-empty command")
	}
	//nolint:gosec // G204: validated session name
	out, err := exec.CommandContext(ctx, s.tmuxBin,
		"respawn-pane", "-k", "-t", s.Name, cmd,
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux respawn-pane: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// SendKeys delivers a single chunk of raw input bytes to the session. Bytes
// are wire-encoded as hex pairs and passed via `tmux send-keys -H`, so escape
// sequences (arrow keys, ctrl combos) round-trip without shell-quoting issues.
func (s *Session) SendKeys(ctx context.Context, raw []byte) error {
	if len(raw) == 0 {
		return nil
	}
	hexArgs := make([]string, 0, sendKeysFixedArgs+len(raw))
	hexArgs = append(hexArgs, "send-keys", "-t", s.Name, "-H")
	for _, b := range raw {
		hexArgs = append(hexArgs, hex.EncodeToString([]byte{b}))
	}
	//nolint:gosec // G204: every hex pair is two bytes of [0-9a-f]
	out, err := exec.CommandContext(ctx, s.tmuxBin, hexArgs...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux send-keys -H: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// CapturePane returns the current pane contents. If includeColors is true the
// output contains tmux's `-e` ANSI sequences; otherwise it is plain text.
func (s *Session) CapturePane(ctx context.Context, includeColors bool) (string, error) {
	args := []string{"capture-pane", "-t", s.Name, "-p"}
	if includeColors {
		args = append(args, "-e")
	}
	//nolint:gosec // G204: subcommand and flags are constants; only session name is variable and validated
	out, err := exec.CommandContext(ctx, s.tmuxBin, args...).CombinedOutput()
	if err != nil {
		return "", fmt.Errorf(
			"tmux capture-pane -t %q: %w (%s)", s.Name, err, strings.TrimSpace(string(out)),
		)
	}
	return string(out), nil
}

// Resize sets the pane size for the session's active window. tmux requires
// resizing the *window*, not the session, for the geometry to take effect.
func (s *Session) Resize(ctx context.Context, width, height int) error {
	if width <= 0 || height <= 0 {
		return fmt.Errorf("tmux resize: width and height must be positive (got %dx%d)", width, height)
	}
	//nolint:gosec // G204: dimensions are integers formatted by strconv
	out, err := exec.CommandContext(ctx, s.tmuxBin,
		"resize-window", "-t", s.Name,
		"-x", strconv.Itoa(width),
		"-y", strconv.Itoa(height),
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux resize-window: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Kill terminates the session unconditionally. Callers should usually go
// through Close, which respects ownership.
func (s *Session) Kill(ctx context.Context) error {
	//nolint:gosec // G204: validated session name
	out, err := exec.CommandContext(ctx, s.tmuxBin, "kill-session", "-t", s.Name).CombinedOutput()
	if err != nil {
		return fmt.Errorf("tmux kill-session: %w (%s)", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// Close kills the session only if this Session created it. Attached existing
// sessions are left running.
func (s *Session) Close(ctx context.Context) error {
	if !s.Owned {
		return nil
	}
	return s.Kill(ctx)
}

func randomSessionName() (string, error) {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("tmux: cannot generate session name: %w", err)
	}
	return "teaeyes-" + hex.EncodeToString(b[:]), nil
}

func validSessionName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9',
			r == '_', r == '-':
		default:
			return false
		}
	}
	return true
}
