package setup

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const tmuxBinding = `bind s display-popup -E -w 60% -h 50% "tmux-pilot"`

// NeedsSetup returns true if tmux-pilot is not yet configured in ~/.tmux.conf.
func NeedsSetup() bool {
	confPath, err := tmuxConfPath()
	if err != nil {
		return false
	}
	exists, err := bindingExists(confPath)
	if err != nil {
		return false
	}
	return !exists
}

// Run adds the tmux-pilot keybinding to ~/.tmux.conf and reloads tmux config.
// It is idempotent — skips if the binding already exists.
func Run() error {
	confPath, err := tmuxConfPath()
	if err != nil {
		return fmt.Errorf("resolve tmux.conf path: %w", err)
	}

	exists, err := bindingExists(confPath)
	if err != nil {
		return err
	}
	if exists {
		fmt.Println("✓ tmux-pilot binding already configured in", confPath)
		return nil
	}

	if err := appendBinding(confPath); err != nil {
		return err
	}

	fmt.Println("✓ Added keybinding to", confPath)
	fmt.Println("  bind s display-popup -E -w 60% -h 50% \"tmux-pilot\"")

	if err := reloadTmux(); err != nil {
		fmt.Println("\n  Reload tmux manually: tmux source-file", confPath)
		return nil // non-fatal: tmux might not be running
	}

	fmt.Println("✓ Reloaded tmux config")
	fmt.Println("\n  Press prefix + s to launch tmux-pilot")
	return nil
}

func tmuxConfPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".tmux.conf"), nil
}

func bindingExists(path string) (bool, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("read %s: %w", path, err)
	}
	return strings.Contains(string(data), "tmux-pilot"), nil
}

func appendBinding(path string) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("open %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	// Add newline before binding if file is non-empty
	info, _ := f.Stat()
	prefix := ""
	if info.Size() > 0 {
		prefix = "\n"
	}

	if _, err := fmt.Fprintf(f, "%s# tmux-pilot: session manager popup\n%s\n", prefix, tmuxBinding); err != nil {
		return fmt.Errorf("write to %s: %w", path, err)
	}
	return nil
}

func reloadTmux() error {
	// Only reload if tmux server is running
	if err := exec.Command("tmux", "list-sessions").Run(); err != nil {
		return err
	}
	home, _ := os.UserHomeDir()
	confPath := filepath.Join(home, ".tmux.conf")
	return exec.Command("tmux", "source-file", confPath).Run()
}
