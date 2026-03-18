package tmux

// Session represents a tmux session with its metadata.
type Session struct {
	Name        string
	WindowCount int
	Attached    bool
}

// ClientOptions configures tmux socket connection.
type ClientOptions struct {
	SocketPath string // -S flag
	SocketName string // -L flag
}

// Args returns tmux arguments for socket configuration.
// Returns -S <path> or -L <name> flags if set.
// If both are set, -S takes precedence.
func (o ClientOptions) Args() []string {
	args := []string{}
	if o.SocketPath != "" {
		args = append(args, "-S", o.SocketPath)
	} else if o.SocketName != "" {
		args = append(args, "-L", o.SocketName)
	}
	return args
}
