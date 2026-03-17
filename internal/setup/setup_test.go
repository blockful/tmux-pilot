package setup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBindingExists_NotPresent(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".tmux.conf")
	if err := os.WriteFile(path, []byte("set -g mouse on\n"), 0644); err != nil {
		t.Fatal(err)
	}

	exists, err := bindingExists(path)
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("should not detect binding in unrelated config")
	}
}

func TestBindingExists_Present(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".tmux.conf")
	if err := os.WriteFile(path, []byte("# tmux-pilot binding\nbind s display-popup -E -w 60% -h 50% \"tmux-pilot\"\n"), 0644); err != nil {
		t.Fatal(err)
	}

	exists, err := bindingExists(path)
	if err != nil {
		t.Fatal(err)
	}
	if !exists {
		t.Error("should detect tmux-pilot in config")
	}
}

func TestBindingExists_FileNotFound(t *testing.T) {
	exists, err := bindingExists("/nonexistent/.tmux.conf")
	if err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("missing file should return false")
	}
}

func TestAppendBinding_NewFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".tmux.conf")

	if err := appendBinding(path); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.Contains(content, "tmux-pilot") {
		t.Error("binding not found in new file")
	}
	if !strings.Contains(content, "display-popup") {
		t.Error("display-popup command not found")
	}
	if strings.HasPrefix(content, "\n") {
		t.Error("new file should not start with newline")
	}
}

func TestAppendBinding_ExistingFile(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".tmux.conf")
	if err := os.WriteFile(path, []byte("set -g mouse on\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := appendBinding(path); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(path)
	content := string(data)

	if !strings.HasPrefix(content, "set -g mouse on\n") {
		t.Error("existing content should be preserved")
	}
	if !strings.Contains(content, "tmux-pilot") {
		t.Error("binding not appended")
	}
}

func TestAppendBinding_Idempotent(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, ".tmux.conf")

	if err := appendBinding(path); err != nil {
		t.Fatal(err)
	}
	first, _ := os.ReadFile(path)

	// Once in comment, once in command
	if count := strings.Count(string(first), "tmux-pilot"); count != 2 {
		t.Errorf("expected 2 occurrences of tmux-pilot, got %d", count)
	}
}
