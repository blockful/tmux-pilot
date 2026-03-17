package tmux

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

// RealClient implements Client by shelling out to the tmux binary.
type RealClient struct{}

// NewRealClient creates a new RealClient.
func NewRealClient() *RealClient {
	return &RealClient{}
}

// ListSessions returns all tmux sessions. Returns an empty slice if the
// tmux server is not running.
func (c *RealClient) ListSessions() ([]Session, error) {
	cmd := exec.Command("tmux", "list-sessions", "-F", "#{session_name}\t#{session_windows}\t#{session_attached}")
	output, err := cmd.Output()
	if err != nil {
		if isServerNotRunning(err) {
			return []Session{}, nil
		}
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	return parseSessions(string(output))
}

// NewSession creates a detached tmux session with the given name.
func (c *RealClient) NewSession(name string) error {
	if err := exec.Command("tmux", "new-session", "-d", "-s", name).Run(); err != nil {
		return fmt.Errorf("create session %q: %w", name, err)
	}
	return nil
}

// SwitchSession switches to the target session. Uses switch-client when
// inside tmux, attach-session when outside.
func (c *RealClient) SwitchSession(name string) error {
	var cmd *exec.Cmd
	if c.IsInsideTmux() {
		cmd = exec.Command("tmux", "switch-client", "-t", name)
	} else {
		cmd = exec.Command("tmux", "attach-session", "-t", name)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("switch to session %q: %w", name, err)
	}
	return nil
}

// RenameSession renames an existing session.
func (c *RealClient) RenameSession(old, new string) error {
	if err := exec.Command("tmux", "rename-session", "-t", old, new).Run(); err != nil {
		return fmt.Errorf("rename session %q to %q: %w", old, new, err)
	}
	return nil
}

// KillSession kills a tmux session.
func (c *RealClient) KillSession(name string) error {
	if err := exec.Command("tmux", "kill-session", "-t", name).Run(); err != nil {
		return fmt.Errorf("kill session %q: %w", name, err)
	}
	return nil
}

// IsInsideTmux returns true if the current process is running inside tmux.
func (c *RealClient) IsInsideTmux() bool {
	return os.Getenv("TMUX") != ""
}

// parseSessions parses tmux list-sessions output (tab-delimited).
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

		windowCount, err := strconv.Atoi(parts[1])
		if err != nil {
			return nil, fmt.Errorf("parse window count %q: %w", parts[1], err)
		}

		sessions = append(sessions, Session{
			Name:        parts[0],
			WindowCount: windowCount,
			Attached:    parts[2] == "1",
		})
	}
	return sessions, nil
}

// isServerNotRunning checks if the error indicates no tmux server is running.
func isServerNotRunning(err error) bool {
	exitError, ok := err.(*exec.ExitError)
	if !ok {
		return false
	}
	stderr := string(exitError.Stderr)
	return strings.Contains(stderr, "no server running") ||
		strings.Contains(stderr, "failed to connect to server") ||
		exitError.ExitCode() == 1
}
