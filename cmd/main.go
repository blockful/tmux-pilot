package main

import (
	"fmt"
	"os"

	"github.com/blockful/tmux-pilot/internal/setup"
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
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--version", "-v":
			fmt.Printf("tmux-pilot %s (%s, %s)\n", version, commit, date)
			os.Exit(0)
		case "--setup":
			if err := setup.Run(); err != nil {
				fmt.Fprintf(os.Stderr, "Setup failed: %v\n", err)
				os.Exit(1)
			}
			os.Exit(0)
		}
	}

	client := tmux.NewRealClient()
	model := tui.New(client, setup.NeedsSetup(), setup.Run)

	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
