package tui

import (
	"fmt"
	"os"

	"github.com/blockful/tmux-pilot/internal/tmux"
)

// Mode represents the current UI state.
type Mode int

const (
	ModeList        Mode = iota
	ModeCreate           // typing new session name
	ModeRename           // typing new name for existing session
	ModeConfirmKill      // y/n confirmation for kill
)

// Action is what the TUI tells the caller to do after exit.
type Action struct {
	Kind    string // "switch", "new", "rename", "kill", "detach", ""
	Target  string // session name
	NewName string // for rename
}

// TUIOptions configures the TUI appearance and behavior.
type TUIOptions struct {
	ColorEnabled bool
}

// State holds the current TUI state.
type State struct {
	sessions     []tmux.Session
	cursor       int
	mode         Mode
	input        string
	warning      string
	action       Action
	clientOpts   tmux.ClientOptions
	terminal     *Terminal
	renderer     *Renderer
	done         bool
}

// Run starts the TUI and returns the user's chosen action.
func Run(sessions []tmux.Session, colorEnabled bool, clientOpts tmux.ClientOptions) (Action, error) {
	// Detect color mode
	colorMode := ColorDisabled
	if colorEnabled {
		// Check for NO_COLOR environment variable
		if os.Getenv("NO_COLOR") == "" {
			// Check if stdout is a terminal
			if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) != 0 {
				colorMode = ColorEnabled
			}
		}
	}

	state := &State{
		sessions:   sessions,
		cursor:     0,
		mode:       ModeList,
		clientOpts: clientOpts,
		terminal:   NewTerminal(),
		renderer:   NewRenderer(colorMode),
	}

	// Enter raw mode
	if err := state.terminal.EnterRawMode(); err != nil {
		return Action{}, fmt.Errorf("enter raw mode: %w", err)
	}

	// Ensure cleanup: clear UI, show cursor, restore terminal — always.
	defer func() {
		state.renderer.Cleanup()
		_ = state.terminal.Restore()
	}()

	// Signal handler also needs to clean up before exit.
	state.terminal.SetupSignalHandlers(func() {
		state.renderer.Cleanup()
	})

	// Main event loop
	for !state.done {
		// Render current state
		width, _ := state.terminal.Size()
		if width < 50 {
			width = 50
		}
		state.renderer.RenderUI(state.sessions, state.cursor, state.mode, state.input, state.warning, width)

		// Read and handle input
		key, err := ReadKey()
		if err != nil {
			return Action{}, fmt.Errorf("read key: %w", err)
		}

		if err := state.handleKey(key); err != nil {
			return Action{}, fmt.Errorf("handle key: %w", err)
		}
	}

	return state.action, nil
}

// handleKey processes a single keystroke based on current mode.
func (s *State) handleKey(key Key) error {
	switch s.mode {
	case ModeList:
		return s.handleListKey(key)
	case ModeCreate:
		return s.handleInputKey(key, "new")
	case ModeRename:
		return s.handleInputKey(key, "rename")
	case ModeConfirmKill:
		return s.handleConfirmKey(key)
	}
	return nil
}

// handleListKey handles input in list mode.
func (s *State) handleListKey(key Key) error {
	switch key.Type {
	case KeyUp:
		if s.cursor > 0 {
			s.cursor--
		}
	case KeyDown:
		if s.cursor < len(s.sessions)-1 {
			s.cursor++
		}
	case KeyEnter:
		if len(s.sessions) > 0 {
			s.action = Action{Kind: "switch", Target: s.sessions[s.cursor].Name}
			s.done = true
		}
	case KeyEscape:
		s.done = true
	case KeyCtrlC:
		// In raw mode, Ctrl-C is byte 0x03, not SIGINT.
		// Cleanup happens via defer in Run(). Just exit the loop.
		s.done = true
	case KeyRune:
		switch key.Rune {
		case 'q':
			s.done = true
		case 'k':
			if s.cursor > 0 {
				s.cursor--
			}
		case 'j':
			if s.cursor < len(s.sessions)-1 {
				s.cursor++
			}
		case 'n':
			s.mode = ModeCreate
			s.input = ""
			s.warning = ""
		case 'r':
			if len(s.sessions) > 0 {
				s.mode = ModeRename
				s.input = s.sessions[s.cursor].Name
				s.warning = ""
			}
		case 'x':
			if len(s.sessions) > 0 {
				s.mode = ModeConfirmKill
			}
		case 'd':
			s.action = Action{Kind: "detach"}
			s.done = true
		}
	}
	return nil
}

// handleInputKey handles input in create/rename modes.
func (s *State) handleInputKey(key Key, kind string) error {
	switch key.Type {
	case KeyEscape:
		s.mode = ModeList
		s.input = ""
		s.warning = ""
	case KeyEnter:
		if s.input == "" {
			return nil
		}
		// Check for duplicate names
		if tmux.SessionExists(s.input, s.clientOpts) {
			s.warning = "'" + s.input + "' already exists"
			return nil
		}
		s.warning = ""
		if kind == "new" {
			s.action = Action{Kind: "new", Target: s.input}
		} else {
			s.action = Action{Kind: "rename", Target: s.sessions[s.cursor].Name, NewName: s.input}
		}
		s.done = true
	case KeyBackspace:
		s.warning = ""
		if len(s.input) > 0 {
			s.input = s.input[:len(s.input)-1]
		}
	case KeyCtrlC:
		s.done = true
	case KeyRune:
		// Accept printable characters
		if key.Rune >= ' ' && key.Rune <= '~' {
			s.warning = ""
			s.input += string(key.Rune)
		}
	}
	return nil
}

// handleConfirmKey handles input in kill confirmation mode.
func (s *State) handleConfirmKey(key Key) error {
	switch key.Type {
	case KeyEnter:
		s.action = Action{Kind: "kill", Target: s.sessions[s.cursor].Name}
		s.done = true
	case KeyEscape:
		s.mode = ModeList
	case KeyCtrlC:
		s.done = true
	case KeyRune:
		switch key.Rune {
		case 'y', 'Y':
			s.action = Action{Kind: "kill", Target: s.sessions[s.cursor].Name}
			s.done = true
		case 'n', 'N', 'q':
			s.mode = ModeList
		}
	}
	return nil
}