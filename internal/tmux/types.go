package tmux

// Session represents a tmux session with its metadata.
type Session struct {
	Name        string
	WindowCount int
	Attached    bool
}
