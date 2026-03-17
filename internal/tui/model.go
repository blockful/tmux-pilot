package tui

import (
	"github.com/blockful/tmux-pilot/internal/tmux"
	tea "github.com/charmbracelet/bubbletea"
)

// Action is what the TUI tells the caller to do after exit.
type Action struct {
	Kind    string // "switch", "new", "rename", "kill", "detach", ""
	Target  string // session name
	NewName string // for rename
}

// Mode is the current UI state.
type Mode int

const (
	ModeList        Mode = iota
	ModeCreate           // typing new session name
	ModeRename           // typing new name
	ModeConfirmKill      // y/n confirmation
)

// Model is the BubbleTea model. It collects user intent and exits.
type Model struct {
	sessions []tmux.Session
	cursor   int
	mode     Mode
	input    string
	warning  string
	width    int
	height   int
	action   Action // populated on exit
}

// New creates a new model with the given sessions.
func New(sessions []tmux.Session) *Model {
	return &Model{
		sessions: sessions,
		width:    80,
		height:   24,
	}
}

// Action returns the action chosen by the user (check after Run).
func (m *Model) Action() Action { return m.action }

// Init is a no-op — sessions are passed in, no async needed.
func (m *Model) Init() tea.Cmd { return nil }

// Update handles input.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch m.mode {
	case ModeList:
		return m.listKey(msg)
	case ModeCreate:
		return m.inputKey(msg, "new")
	case ModeRename:
		return m.inputKey(msg, "rename")
	case ModeConfirmKill:
		return m.confirmKey(msg)
	}
	return m, nil
}

func (m *Model) listKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		if m.cursor > 0 {
			m.cursor--
		}
		return m, nil
	case tea.KeyDown:
		if m.cursor < len(m.sessions)-1 {
			m.cursor++
		}
		return m, nil
	case tea.KeyEnter:
		if len(m.sessions) > 0 {
			m.action = Action{Kind: "switch", Target: m.sessions[m.cursor].Name}
			return m, tea.Quit
		}
		return m, nil
	case tea.KeyEscape:
		return m, tea.Quit
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "j":
		if m.cursor < len(m.sessions)-1 {
			m.cursor++
		}
	case "n":
		m.mode = ModeCreate
		m.input = ""
		m.warning = ""
	case "r":
		if len(m.sessions) > 0 {
			m.mode = ModeRename
			m.input = m.sessions[m.cursor].Name
			m.warning = ""
		}
	case "x":
		if len(m.sessions) > 0 {
			m.mode = ModeConfirmKill
		}
	case "d":
		m.action = Action{Kind: "detach"}
		return m, tea.Quit
	}
	return m, nil
}

func (m *Model) inputKey(msg tea.KeyMsg, kind string) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		m.mode = ModeList
		m.input = ""
		m.warning = ""
		return m, nil
	case tea.KeyEnter:
		if m.input == "" {
			return m, nil
		}
		// Check duplicate
		if tmux.SessionExists(m.input) {
			m.warning = "'" + m.input + "' already exists"
			return m, nil
		}
		m.warning = ""
		if kind == "new" {
			m.action = Action{Kind: "new", Target: m.input}
		} else {
			m.action = Action{Kind: "rename", Target: m.sessions[m.cursor].Name, NewName: m.input}
		}
		return m, tea.Quit
	case tea.KeyBackspace:
		m.warning = ""
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil
	}

	key := msg.String()
	if len(key) == 1 && key[0] >= ' ' && key[0] <= '~' {
		m.warning = ""
		m.input += key
	}
	return m, nil
}

func (m *Model) confirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y":
		m.action = Action{Kind: "kill", Target: m.sessions[m.cursor].Name}
		return m, tea.Quit
	case "n", "q":
		m.mode = ModeList
		return m, nil
	}
	switch msg.Type {
	case tea.KeyEnter:
		m.action = Action{Kind: "kill", Target: m.sessions[m.cursor].Name}
		return m, tea.Quit
	case tea.KeyEscape:
		m.mode = ModeList
		return m, nil
	}
	return m, nil
}

// --- Getters for view ---

func (m *Model) Sessions() []tmux.Session { return m.sessions }
func (m *Model) Cursor() int              { return m.cursor }
func (m *Model) Mode() Mode               { return m.mode }
func (m *Model) Input() string             { return m.input }
func (m *Model) Warning() string           { return m.warning }
func (m *Model) Width() int                { return m.width }
func (m *Model) KillName() string {
	if m.mode == ModeConfirmKill && m.cursor < len(m.sessions) {
		return m.sessions[m.cursor].Name
	}
	return ""
}
