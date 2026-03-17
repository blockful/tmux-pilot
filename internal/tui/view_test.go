package tui

import (
	"strings"
	"testing"
)

func TestView_List(t *testing.T) {
	m := newTestModel()
	v := m.View()
	for _, want := range []string{"tmux-pilot", "main", "api", "notes", "[enter] switch", "[d] detach"} {
		if !strings.Contains(v, want) {
			t.Errorf("list view missing %q", want)
		}
	}
}

func TestView_Empty(t *testing.T) {
	m := New(nil)
	if !strings.Contains(m.View(), "No sessions") {
		t.Error("empty view should say 'No sessions'")
	}
}

func TestView_Create(t *testing.T) {
	m := update(newTestModel(), key("n"), key("d"), key("e"), key("v"))
	v := m.View()
	if !strings.Contains(v, "New session name:") || !strings.Contains(v, "dev") {
		t.Error("create view missing prompt or input")
	}
}

func TestView_Confirm(t *testing.T) {
	m := update(newTestModel(), key("x"))
	if !strings.Contains(m.View(), "Kill session 'main'?") {
		t.Error("confirm view missing session name")
	}
}

func TestView_Warning(t *testing.T) {
	m := update(newTestModel(), key("n"))
	m.warning = "already exists"
	if !strings.Contains(m.View(), "already exists") {
		t.Error("warning not shown")
	}
}
