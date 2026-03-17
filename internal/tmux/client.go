package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// ListSessions returns all tmux sessions.
func ListSessions() ([]Session, error) {
	out, err := exec.Command("tmux", "list-sessions", "-F", "#{session_name}\t#{session_windows}\t#{session_attached}").Output()
	if err != nil {
		if isNoServer(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	return parseSessions(string(out))
}

// SwitchOrAttach switches to a session (inside tmux) or attaches (outside).
func SwitchOrAttach(name string) error {
	if IsInsideTmux() {
		return run("switch-client", "-t", name)
	}
	cmd := exec.Command("tmux", "attach-session", "-t", name)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// NewSession creates a session and switches to it.
func NewSession(name string) error {
	if err := run("new-session", "-d", "-s", name); err != nil {
		return err
	}
	return SwitchOrAttach(name)
}

// RenameSession renames a session.
func RenameSession(old, new string) error {
	return run("rename-session", "-t", old, new)
}

// KillSession kills a session.
func KillSession(name string) error {
	return run("kill-session", "-t", name)
}

// Detach detaches the current client.
func Detach() error {
	return run("detach-client")
}

// IsInsideTmux returns true if running inside a tmux session.
func IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

// SessionExists checks if a session name is already taken.
func SessionExists(name string) bool {
	sessions, err := ListSessions()
	if err != nil {
		return false
	}
	for _, s := range sessions {
		if s.Name == name {
			return true
		}
	}
	return false
}

func run(args ...string) error {
	return exec.Command("tmux", args...).Run()
}

func parseSessions(output string) ([]Session, error) {
	var sessions []Session
	for _, line := range strings.Split(strings.TrimSpace(output), "\n") {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, "\t", 3)
		if len(parts) != 3 {
			continue
		}
		wc, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("parse window count %q: %w", parts[1], err)
		}
		sessions = append(sessions, Session{
			Name:        parts[0],
			WindowCount: wc,
			Attached:    parts[2] == "1",
		})
	}
	return sessions, nil
}

func isNoServer(err error) bool {
	if e, ok := err.(*exec.ExitError); ok {
		stderr := string(e.Stderr)
		return strings.Contains(stderr, "no server running") ||
			strings.Contains(stderr, "failed to connect") ||
			e.ExitCode() == 1
	}
	return false
}
