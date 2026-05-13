// Command tea-eyes is the MCP server entry point for the tea-eyes plugin.
//
// It speaks the Model Context Protocol over stdio and exposes tools that let
// Claude Code see, drive, and test terminal user interfaces.
//
// Subcommands:
//
//	tea-eyes              run the MCP server on stdio (default)
//	tea-eyes doctor       check for required external binaries (vhs, ttyd, ffmpeg)
//	tea-eyes cache clean  remove cached renders
//	tea-eyes -version     print version and exit
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	mcpserver "github.com/mark3labs/mcp-go/server"

	"gitlab.com/skjutare/tea-eyes/internal/render"
	"gitlab.com/skjutare/tea-eyes/internal/server"
)

// exitCodeUsage is returned for CLI usage errors (missing/unknown subcommand).
const exitCodeUsage = 2

func main() {
	args := os.Args[1:]
	if len(args) > 0 {
		switch args[0] {
		case "doctor":
			os.Exit(runDoctor())
		case "cache":
			os.Exit(runCache(args[1:]))
		case "serve":
			os.Exit(runServer(args[1:]))
		}
	}
	os.Exit(runServer(args))
}

func runServer(args []string) int {
	fs := flag.NewFlagSet("tea-eyes", flag.ExitOnError)
	showVersion := fs.Bool("version", false, "print version and exit")
	logFile := fs.String("log-file", "", "append logs to this file (default: stderr)")
	_ = fs.Parse(args)

	if *showVersion {
		fmt.Println("tea-eyes", server.Version)
		return 0
	}

	var logOut io.Writer = os.Stderr
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tea-eyes: cannot open log file %q: %v\n", *logFile, err)
			return 1
		}
		defer func() { _ = f.Close() }()
		logOut = f
	}
	log.SetOutput(logOut)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("tea-eyes %s starting", server.Version)

	s := server.New()
	if err := mcpserver.ServeStdio(s); err != nil {
		log.Printf("tea-eyes: server exited with error: %v", err)
		return 1
	}
	return 0
}

// runDoctor checks for vhs/ttyd/ffmpeg on PATH and prints versions or install
// hints. Returns the process exit code: 0 if all present, 1 if anything is
// missing.
func runDoctor() int {
	type check struct {
		name     string
		args     []string
		hint     string
		optional bool
	}
	checks := []check{
		{"vhs", []string{"--version"},
			"install with `brew install vhs` or `go install github.com/charmbracelet/vhs@latest`", false},
		{"ttyd", []string{"--version"},
			"install with `brew install ttyd` (required by vhs)", false},
		{"ffmpeg", []string{"-version"},
			"install with `brew install ffmpeg` (required by vhs)", false},
		{"tmux", []string{"-V"},
			"install with `brew install tmux` (optional — only required for mode=\"tmux\")", true},
	}
	missingRequired := 0
	missingOptional := 0
	fmt.Printf("tea-eyes %s — doctor\n\n", server.Version)
	for _, c := range checks {
		path, err := exec.LookPath(c.name)
		if err != nil {
			tag := "✗"
			label := "not found on PATH"
			if c.optional {
				tag = "•"
				label = "not found on PATH (optional)"
				missingOptional++
			} else {
				missingRequired++
			}
			fmt.Printf("  %s %-8s %s\n      %s\n", tag, c.name, label, c.hint)
			continue
		}
		ver := firstLine(captureVersion(path, c.args))
		fmt.Printf("  ✓ %-8s %s  (%s)\n", c.name, ver, path)
	}
	cache := render.DefaultCacheDir()
	fmt.Printf("\n  cache dir: %s\n", cache)
	if missingRequired > 0 {
		fmt.Printf(
			"\n%d required dependency(ies) missing. tui_render_image will fail until they are installed.\n",
			missingRequired,
		)
		return 1
	}
	if missingOptional > 0 {
		fmt.Printf(
			"\nAll required dependencies present. %d optional dependency(ies) missing — affected features noted above.\n",
			missingOptional,
		)
		return 0
	}
	fmt.Println("\nAll dependencies present.")
	return 0
}

// runCache implements `tea-eyes cache <subcommand>`.
func runCache(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: tea-eyes cache <clean|path>")
		return exitCodeUsage
	}
	switch args[0] {
	case "clean":
		r := render.NewRenderer("")
		n, err := r.CleanCache()
		if err != nil {
			fmt.Fprintf(os.Stderr, "tea-eyes cache clean: %v\n", err)
			return 1
		}
		fmt.Printf("removed %d cached file(s) from %s\n", n, r.CacheDir())
		return 0
	case "path":
		fmt.Println(render.DefaultCacheDir())
		return 0
	default:
		//nolint:gosec // G705: writing to stderr is not an HTTP response; XSS is a false positive
		fmt.Fprintf(os.Stderr, "tea-eyes cache: unknown subcommand %q (want clean|path)\n", args[0])
		return exitCodeUsage
	}
}

func captureVersion(bin string, args []string) string {
	out, err := exec.CommandContext(context.Background(), bin, args...).CombinedOutput()
	if err != nil && len(out) == 0 {
		return "(version check failed: " + err.Error() + ")"
	}
	return strings.TrimSpace(string(out))
}

func firstLine(s string) string {
	if before, _, ok := strings.Cut(s, "\n"); ok {
		return before
	}
	return s
}
