package main

import (
	"fmt"
	"os"

	"github.com/blockful/tmux-pilot/internal/tmux"
	"github.com/blockful/tmux-pilot/internal/tui"
	tea "github.com/charmbracelet/bubbletea"
)

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
			return
		}
	}

	// 1. Fetch sessions
	sessions, err := tmux.ListSessions()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 2. Show picker
	model := tui.New(sessions)
	p := tea.NewProgram(model, tea.WithAltScreen(), tea.WithInputTTY())
	result, err := p.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// 3. Execute action after TUI exits
	action := result.(*tui.Model).Action()
	if err := execute(action); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func execute(a tui.Action) error {
	switch a.Kind {
	case "switch":
		return tmux.SwitchOrAttach(a.Target)
	case "new":
		return tmux.NewSession(a.Target)
	case "rename":
		return tmux.RenameSession(a.Target, a.NewName)
	case "kill":
		return tmux.KillSession(a.Target)
	case "detach":
		return tmux.Detach()
	}
	return nil // user quit without action
}
