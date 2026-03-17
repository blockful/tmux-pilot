package main

import (
	"fmt"
	"os"

	"github.com/blockful/tmux-pilot/internal/tmux"
	"github.com/blockful/tmux-pilot/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

// Set by goreleaser ldflags.
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "-v") {
		fmt.Printf("tmux-pilot %s (%s, %s)\n", version, commit, date)
		os.Exit(0)
	}

	client := tmux.NewRealClient()
	model := tui.New(client)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
