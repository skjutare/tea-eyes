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

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "doctor":
			os.Exit(runDoctor())
		case "cache":
			os.Exit(runCache(os.Args[2:]))
		}
	}
	runServer()
}

func runServer() {
	var (
		showVersion = flag.Bool("version", false, "print version and exit")
		logFile     = flag.String("log-file", "", "append logs to this file (default: stderr)")
	)
	flag.Parse()

	if *showVersion {
		fmt.Println("tea-eyes", server.Version)
		return
	}

	var logOut io.Writer = os.Stderr
	if *logFile != "" {
		f, err := os.OpenFile(*logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tea-eyes: cannot open log file %q: %v\n", *logFile, err)
			os.Exit(1)
		}
		defer f.Close()
		logOut = f
	}
	log.SetOutput(logOut)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)
	log.Printf("tea-eyes %s starting", server.Version)

	s := server.New()
	if err := mcpserver.ServeStdio(s); err != nil {
		log.Printf("tea-eyes: server exited with error: %v", err)
		os.Exit(1)
	}
}

// runDoctor checks for vhs/ttyd/ffmpeg on PATH and prints versions or install
// hints. Returns the process exit code: 0 if all present, 1 if anything is
// missing.
func runDoctor() int {
	checks := []struct {
		name string
		args []string
		hint string
	}{
		{"vhs", []string{"--version"},
			"install with `brew install vhs` or `go install github.com/charmbracelet/vhs@latest`"},
		{"ttyd", []string{"--version"},
			"install with `brew install ttyd` (required by vhs)"},
		{"ffmpeg", []string{"-version"},
			"install with `brew install ffmpeg` (required by vhs)"},
	}
	missing := 0
	fmt.Printf("tea-eyes %s — doctor\n\n", server.Version)
	for _, c := range checks {
		path, err := exec.LookPath(c.name)
		if err != nil {
			fmt.Printf("  ✗ %-8s not found on PATH\n      %s\n", c.name, c.hint)
			missing++
			continue
		}
		ver := firstLine(captureVersion(path, c.args))
		fmt.Printf("  ✓ %-8s %s  (%s)\n", c.name, ver, path)
	}
	cache := render.DefaultCacheDir()
	fmt.Printf("\n  cache dir: %s\n", cache)
	if missing > 0 {
		fmt.Printf("\n%d external dependency(ies) missing. tui_render_image will fail until they are installed.\n", missing)
		return 1
	}
	fmt.Println("\nAll external dependencies present.")
	return 0
}

// runCache implements `tea-eyes cache <subcommand>`.
func runCache(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: tea-eyes cache <clean|path>")
		return 2
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
		fmt.Fprintf(os.Stderr, "tea-eyes cache: unknown subcommand %q (want clean|path)\n", args[0])
		return 2
	}
}

func captureVersion(bin string, args []string) string {
	out, err := exec.Command(bin, args...).CombinedOutput()
	if err != nil && len(out) == 0 {
		return "(version check failed: " + err.Error() + ")"
	}
	return strings.TrimSpace(string(out))
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}
