package tui

import (
	"github.com/blockful/tmux-pilot/internal/tmux"
	tea "github.com/charmbracelet/bubbletea"
)

// Mode represents the current UI mode.
type Mode int

const (
	ModeList        Mode = iota // Browsing sessions
	ModeCreate                  // Typing a new session name
	ModeRename                  // Typing a new name for existing session
	ModeConfirmKill             // Confirming session kill
	ModeSetup                   // First-run setup prompt
)

// sessionsMsg carries a refreshed session list.
type sessionsMsg struct {
	sessions []tmux.Session
	err      error
}

// operationMsg signals that an async tmux operation completed.
type operationMsg struct {
	err      error
	switchTo string // if non-empty, quit after switching
}

// setupDoneMsg signals that the setup operation completed.
type setupDoneMsg struct {
	err error
}

// SetupFunc is the function called to perform setup.
// Injected to keep the model testable without filesystem side effects.
type SetupFunc func() error

// Model is the BubbleTea model for tmux-pilot.
type Model struct {
	client    tmux.Client
	sessions  []tmux.Session
	cursor    int
	mode      Mode
	input     string
	killName  string // session name pending kill confirmation
	err       error
	warning   string // non-fatal warning (e.g. duplicate name)
	width     int
	height    int
	quitting  bool
	setupFunc SetupFunc
}

// New creates a Model wired to the given tmux client.
// If needsSetup is true, the first screen will prompt to configure tmux.
func New(client tmux.Client, needsSetup bool, setupFn SetupFunc) *Model {
	mode := ModeList
	if needsSetup {
		mode = ModeSetup
	}
	return &Model{
		client:    client,
		width:     80,
		height:    24,
		mode:      mode,
		setupFunc: setupFn,
	}
}

// Init fetches the initial session list.
func (m *Model) Init() tea.Cmd {
	return m.fetchSessions()
}

// Update processes messages and returns the updated model + next command.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case sessionsMsg:
		m.sessions = msg.sessions
		m.err = msg.err
		m.clampCursor()
		return m, nil

	case operationMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		m.mode = ModeList
		m.input = ""
		m.killName = ""
		m.warning = ""
		if msg.switchTo != "" {
			m.quitting = true
			return m, tea.Quit
		}
		return m, m.fetchSessions()

	case setupDoneMsg:
		if msg.err != nil {
			m.err = msg.err
		}
		m.mode = ModeList
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	return m, nil
}

// handleKey dispatches key presses to the current mode handler.
func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Normalize key string. BubbleTea converts both ESC[A and ESCOA
	// to "up", but we also handle KeyType for robustness.
	key := msg.String()

	switch m.mode {
	case ModeSetup:
		return m.handleSetupKey(key)
	case ModeList:
		return m.handleListKey(msg)
	case ModeCreate:
		return m.handleInputKey(key, m.doCreate)
	case ModeRename:
		return m.handleInputKey(key, m.doRename)
	case ModeConfirmKill:
		return m.handleConfirmKey(key)
	}
	return m, nil
}

// handleSetupKey handles keys in the first-run setup prompt.
func (m *Model) handleSetupKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "y", "enter":
		return m, m.runSetup()
	case "n", "esc", "q":
		m.mode = ModeList
		return m, nil
	}
	return m, nil
}

// handleListKey handles keys in the session list view.
func (m *Model) handleListKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Use KeyType for special keys (more robust than string matching)
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
			name := m.sessions[m.cursor].Name
			return m, m.switchSession(name)
		}
		return m, nil
	case tea.KeyEscape:
		m.quitting = true
		return m, tea.Quit
	}

	// Use string for single-char keys
	switch msg.String() {
	case "q":
		m.quitting = true
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
		m.err = nil
	case "r":
		if len(m.sessions) > 0 {
			m.mode = ModeRename
			m.input = m.sessions[m.cursor].Name
			m.err = nil
		}
	case "x":
		if len(m.sessions) > 0 {
			m.mode = ModeConfirmKill
			m.killName = m.sessions[m.cursor].Name
			m.err = nil
		}
	case "d":
		return m, m.detachSession()
	}
	return m, nil
}

// handleInputKey handles keys in create/rename modes.
// submit is called when enter is pressed with non-empty input.
func (m *Model) handleInputKey(key string, submit func() tea.Cmd) (tea.Model, tea.Cmd) {
	// Clear warning on any input change
	switch key {
	case "esc":
		m.mode = ModeList
		m.input = ""
		m.warning = ""
		return m, nil
	case "enter":
		if m.input != "" {
			if m.sessionNameExists(m.input) {
				m.warning = "Session '" + m.input + "' already exists"
				return m, nil
			}
			m.warning = ""
			return m, submit()
		}
		return m, nil
	case "backspace":
		m.warning = ""
		if len(m.input) > 0 {
			m.input = m.input[:len(m.input)-1]
		}
		return m, nil
	default:
		m.warning = ""
		if len(key) == 1 && key[0] >= ' ' && key[0] <= '~' {
			m.input += key
		}
		return m, nil
	}
}

// sessionNameExists checks if a session name is already in use.
func (m *Model) sessionNameExists(name string) bool {
	for _, s := range m.sessions {
		if s.Name == name {
			return true
		}
	}
	return false
}

// handleConfirmKey handles keys in the kill confirmation dialog.
func (m *Model) handleConfirmKey(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "y", "enter":
		name := m.killName
		return m, m.killSession(name)
	case "n", "esc":
		m.mode = ModeList
		m.killName = ""
		return m, nil
	}
	return m, nil
}

// clampCursor ensures the cursor stays within bounds.
func (m *Model) clampCursor() {
	if len(m.sessions) == 0 {
		m.cursor = 0
	} else if m.cursor >= len(m.sessions) {
		m.cursor = len(m.sessions) - 1
	}
}

// --- Async commands ---

func (m *Model) fetchSessions() tea.Cmd {
	return func() tea.Msg {
		sessions, err := m.client.ListSessions()
		return sessionsMsg{sessions: sessions, err: err}
	}
}

func (m *Model) doCreate() tea.Cmd {
	name := m.input
	return func() tea.Msg {
		if err := m.client.NewSession(name); err != nil {
			return operationMsg{err: err}
		}
		return operationMsg{switchTo: name}
	}
}

func (m *Model) doRename() tea.Cmd {
	oldName := m.sessions[m.cursor].Name
	newName := m.input
	return func() tea.Msg {
		err := m.client.RenameSession(oldName, newName)
		return operationMsg{err: err}
	}
}

func (m *Model) switchSession(name string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.SwitchSession(name)
		return operationMsg{err: err, switchTo: name}
	}
}

func (m *Model) killSession(name string) tea.Cmd {
	return func() tea.Msg {
		err := m.client.KillSession(name)
		return operationMsg{err: err}
	}
}

func (m *Model) detachSession() tea.Cmd {
	return func() tea.Msg {
		err := m.client.DetachSession()
		return operationMsg{err: err, switchTo: "detach"}
	}
}

func (m *Model) runSetup() tea.Cmd {
	return func() tea.Msg {
		if m.setupFunc == nil {
			return setupDoneMsg{}
		}
		err := m.setupFunc()
		return setupDoneMsg{err: err}
	}
}

// --- Getters for view and testing ---

// Mode returns the current UI mode.
func (m *Model) Mode() Mode { return m.mode }

// Sessions returns the current session list.
func (m *Model) Sessions() []tmux.Session { return m.sessions }

// Cursor returns the current cursor position.
func (m *Model) Cursor() int { return m.cursor }

// Input returns the current input buffer.
func (m *Model) Input() string { return m.input }

// KillName returns the name of the session pending kill confirmation.
func (m *Model) KillName() string { return m.killName }

// Err returns the current error, if any.
func (m *Model) Err() error { return m.err }

// Warning returns the current warning message, if any.
func (m *Model) Warning() string { return m.warning }

// Width returns the terminal width.
func (m *Model) Width() int { return m.width }

// IsQuitting returns true if the model is in quitting state.
func (m *Model) IsQuitting() bool { return m.quitting }
