// Command tea-eyes is the MCP server entry point for the tea-eyes plugin.
//
// It speaks the Model Context Protocol over stdio and exposes tools that let
// Claude Code see, drive, and test terminal user interfaces.
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	mcpserver "github.com/mark3labs/mcp-go/server"

	"gitlab.com/skjutare/tea-eyes/internal/server"
)

func main() {
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
