package tui

import (
	"strings"
	"testing"

	"github.com/blockful/tmux-pilot/internal/tmux"
)

func TestView_ListMode(t *testing.T) {
	m, _ := testModel()
	view := m.View()

	checks := []string{
		"tmux-pilot",
		"main",
		"api-server",
		"notes",
		"attached",
		"detached",
		"[enter] switch",
		"[n] new",
		"[r] rename",
		"[x] kill",
		"[q] quit",
		"Ctrl-b d",
	}
	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("list view missing %q", check)
		}
	}
}

func TestView_ListMode_Empty(t *testing.T) {
	m, _ := testModel()
	m.sessions = nil
	view := m.View()

	if !strings.Contains(view, "No sessions found") {
		t.Error("empty list should show 'No sessions found'")
	}
}

func TestView_ListMode_WindowPluralization(t *testing.T) {
	m, _ := testModel()
	m.sessions = []tmux.Session{
		{Name: "one", WindowCount: 1, Attached: false},
		{Name: "many", WindowCount: 5, Attached: false},
	}
	view := m.View()

	if !strings.Contains(view, "1 window ") {
		t.Error("single window should not be pluralized")
	}
	if !strings.Contains(view, "5 windows") {
		t.Error("multiple windows should be pluralized")
	}
}

func TestView_ListMode_Indicators(t *testing.T) {
	m, _ := testModel()
	view := m.View()

	if !strings.Contains(view, "●") {
		t.Error("attached session should have ● indicator")
	}
	if !strings.Contains(view, "○") {
		t.Error("detached session should have ○ indicator")
	}
}

func TestView_CreateMode(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeCreate
	m.input = "new-session"
	view := m.View()

	if !strings.Contains(view, "Create new session:") {
		t.Error("create view missing label")
	}
	if !strings.Contains(view, "new-session") {
		t.Error("create view missing input text")
	}
	if !strings.Contains(view, "[enter] create") {
		t.Error("create view missing help")
	}
}

func TestView_RenameMode(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeRename
	m.cursor = 0
	m.input = "primary"
	view := m.View()

	if !strings.Contains(view, "Rename 'main':") {
		t.Error("rename view missing session name")
	}
	if !strings.Contains(view, "primary") {
		t.Error("rename view missing input text")
	}
}

func TestView_ConfirmKill(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeConfirmKill
	m.killName = "notes"
	view := m.View()

	if !strings.Contains(view, "Kill session 'notes'?") {
		t.Error("confirm view missing session name")
	}
	if !strings.Contains(view, "[y/enter] yes") {
		t.Error("confirm view missing help")
	}
}

func TestView_WithError(t *testing.T) {
	m, _ := testModel()
	m.err = &testError{"something broke"}
	view := m.View()

	if !strings.Contains(view, "Error: something broke") {
		t.Error("error should be displayed")
	}
}

func TestView_Quitting(t *testing.T) {
	m, _ := testModel()
	m.quitting = true
	view := m.View()

	if view != "" {
		t.Error("quitting should render empty string")
	}
}

type testError struct{ msg string }

func (e *testError) Error() string { return e.msg }

func TestView_SetupMode(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeSetup
	view := m.View()

	checks := []string{
		"tmux-pilot",
		"Add tmux keybinding",
		"prefix + s",
		".tmux.conf",
		"[y/enter] yes",
		"[n/esc] skip",
	}
	for _, check := range checks {
		if !strings.Contains(view, check) {
			t.Errorf("setup view missing %q", check)
		}
	}
}
