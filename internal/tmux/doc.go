// Package tmux wraps the tmux CLI to drive a TUI inside a persistent or
// ephemeral tmux session. Compared to the pty driver, tmux mode lets the user
// attach to the same session in another terminal and watch Claude work live.
//
// The wrapper shells out to tmux for every operation; there is no long-lived
// connection. tmux itself maintains the session state between calls, so the
// overhead is small compared to the time spent inside the TUI.
package tmux
