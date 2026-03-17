package tmux

import "errors"

// MockClient implements Client for testing.
type MockClient struct {
	sessions   []Session
	insideTmux bool

	ListErr   error
	NewErr    error
	SwitchErr error
	RenameErr error
	KillErr   error
	DetachErr error

	// Recorded calls for verification
	NewCalls    []string
	SwitchCalls []string
	RenameCalls [][2]string
	KillCalls   []string
	DetachCalls int
}

// NewMockClient creates a MockClient with default test sessions.
func NewMockClient() *MockClient {
	return &MockClient{
		sessions: []Session{
			{Name: "main", WindowCount: 3, Attached: true},
			{Name: "api-server", WindowCount: 1, Attached: false},
			{Name: "notes", WindowCount: 2, Attached: false},
		},
		insideTmux: true,
	}
}

func (m *MockClient) ListSessions() ([]Session, error) {
	if m.ListErr != nil {
		return nil, m.ListErr
	}
	return m.sessions, nil
}

func (m *MockClient) NewSession(name string) error {
	m.NewCalls = append(m.NewCalls, name)
	if m.NewErr != nil {
		return m.NewErr
	}
	m.sessions = append(m.sessions, Session{Name: name, WindowCount: 1, Attached: false})
	return nil
}

func (m *MockClient) SwitchSession(name string) error {
	m.SwitchCalls = append(m.SwitchCalls, name)
	if m.SwitchErr != nil {
		return m.SwitchErr
	}
	for i := range m.sessions {
		m.sessions[i].Attached = (m.sessions[i].Name == name)
	}
	return nil
}

func (m *MockClient) RenameSession(old, new string) error {
	m.RenameCalls = append(m.RenameCalls, [2]string{old, new})
	if m.RenameErr != nil {
		return m.RenameErr
	}
	for i := range m.sessions {
		if m.sessions[i].Name == old {
			m.sessions[i].Name = new
			return nil
		}
	}
	return errors.New("session not found: " + old)
}

func (m *MockClient) DetachSession() error {
	m.DetachCalls++
	if m.DetachErr != nil {
		return m.DetachErr
	}
	return nil
}

func (m *MockClient) KillSession(name string) error {
	m.KillCalls = append(m.KillCalls, name)
	if m.KillErr != nil {
		return m.KillErr
	}
	for i, s := range m.sessions {
		if s.Name == name {
			m.sessions = append(m.sessions[:i], m.sessions[i+1:]...)
			return nil
		}
	}
	return errors.New("session not found: " + name)
}

func (m *MockClient) IsInsideTmux() bool {
	return m.insideTmux
}

// SetSessions replaces the mock's session list.
func (m *MockClient) SetSessions(sessions []Session) {
	m.sessions = sessions
}

// SetInsideTmux sets the IsInsideTmux return value.
func (m *MockClient) SetInsideTmux(inside bool) {
	m.insideTmux = inside
}

// Sessions returns the current session list (for test assertions).
func (m *MockClient) Sessions() []Session {
	return m.sessions
}
