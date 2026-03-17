package tmux

// Session represents a tmux session with its metadata.
type Session struct {
	Name        string
	WindowCount int
	Attached    bool
}

// Client defines the interface for tmux operations.
// All operations shell out to the tmux binary.
type Client interface {
	ListSessions() ([]Session, error)
	NewSession(name string) error
	SwitchSession(name string) error
	RenameSession(old, new string) error
	KillSession(name string) error
	DetachSession() error
	IsInsideTmux() bool
}
