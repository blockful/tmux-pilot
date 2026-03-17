package tui

import (
	"testing"

	"github.com/blockful/tmux-pilot/internal/tmux"
	tea "github.com/charmbracelet/bubbletea"
)

var testSessions = []tmux.Session{
	{Name: "main", WindowCount: 3, Attached: true},
	{Name: "api", WindowCount: 1, Attached: false},
	{Name: "notes", WindowCount: 2, Attached: false},
}

func newTestModel() *Model { return New(testSessions) }

func key(k string) tea.KeyMsg   { return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)} }
func special(t tea.KeyType) tea.KeyMsg { return tea.KeyMsg{Type: t} }

func update(m *Model, msgs ...tea.Msg) *Model {
	for _, msg := range msgs {
		result, _ := m.Update(msg)
		m = result.(*Model)
	}
	return m
}

// --- Navigation ---

func TestNavigation(t *testing.T) {
	m := update(newTestModel(), key("j"))
	if m.Cursor() != 1 {
		t.Errorf("j: want cursor 1, got %d", m.Cursor())
	}

	m = update(m, key("k"))
	if m.Cursor() != 0 {
		t.Errorf("k: want cursor 0, got %d", m.Cursor())
	}

	m = update(newTestModel(), special(tea.KeyDown))
	if m.Cursor() != 1 {
		t.Errorf("down: want cursor 1, got %d", m.Cursor())
	}

	m = update(newTestModel(), special(tea.KeyUp))
	if m.Cursor() != 0 {
		t.Error("up at 0 should stay 0")
	}
}

func TestNavigation_Bounds(t *testing.T) {
	m := newTestModel()
	m = update(m, key("j"), key("j"), key("j"), key("j"))
	if m.Cursor() != 2 {
		t.Errorf("should clamp at last index, got %d", m.Cursor())
	}
}

// --- Switch ---

func TestSwitch(t *testing.T) {
	m := newTestModel()
	result, cmd := m.Update(special(tea.KeyEnter))
	model := result.(*Model)
	if model.Action().Kind != "switch" || model.Action().Target != "main" {
		t.Errorf("want switch/main, got %v", model.Action())
	}
	if cmd == nil {
		t.Error("should quit")
	}
}

// --- Quit ---

func TestQuit(t *testing.T) {
	m := newTestModel()
	_, cmd := m.Update(key("q"))
	if cmd == nil {
		t.Error("q should quit")
	}
	if m.Action().Kind != "" {
		t.Error("quit should have no action")
	}
}

func TestQuit_Esc(t *testing.T) {
	_, cmd := newTestModel().Update(special(tea.KeyEscape))
	if cmd == nil {
		t.Error("esc should quit")
	}
}

// --- Create mode ---

func TestCreate(t *testing.T) {
	m := update(newTestModel(), key("n"))
	if m.Mode() != ModeCreate {
		t.Error("n should enter create mode")
	}

	m = update(m, key("d"), key("e"), key("v"))
	if m.Input() != "dev" {
		t.Errorf("want input 'dev', got %q", m.Input())
	}

	result, cmd := m.Update(special(tea.KeyEnter))
	model := result.(*Model)
	if model.Action().Kind != "new" || model.Action().Target != "dev" {
		t.Errorf("want new/dev, got %v", model.Action())
	}
	if cmd == nil {
		t.Error("should quit after create")
	}
}

func TestCreate_Cancel(t *testing.T) {
	m := update(newTestModel(), key("n"), key("d"))
	m = update(m, special(tea.KeyEscape))
	if m.Mode() != ModeList {
		t.Error("esc should return to list")
	}
}

func TestCreate_Backspace(t *testing.T) {
	m := update(newTestModel(), key("n"), key("a"), key("b"))
	m = update(m, special(tea.KeyBackspace))
	if m.Input() != "a" {
		t.Errorf("want 'a', got %q", m.Input())
	}
}

func TestCreate_EmptyEnter(t *testing.T) {
	m := update(newTestModel(), key("n"))
	_, cmd := m.Update(special(tea.KeyEnter))
	if cmd != nil {
		t.Error("empty enter should not quit")
	}
}

// --- Rename ---

func TestRename(t *testing.T) {
	m := update(newTestModel(), key("r"))
	if m.Mode() != ModeRename {
		t.Error("r should enter rename mode")
	}
	if m.Input() != "main" {
		t.Errorf("should prefill with 'main', got %q", m.Input())
	}
}

// --- Kill ---

func TestKill(t *testing.T) {
	m := update(newTestModel(), key("x"))
	if m.Mode() != ModeConfirmKill {
		t.Error("x should enter confirm mode")
	}

	result, cmd := m.Update(key("y"))
	model := result.(*Model)
	if model.Action().Kind != "kill" || model.Action().Target != "main" {
		t.Errorf("want kill/main, got %v", model.Action())
	}
	if cmd == nil {
		t.Error("should quit after kill")
	}
}

func TestKill_Cancel(t *testing.T) {
	m := update(newTestModel(), key("x"))
	m = update(m, key("n"))
	if m.Mode() != ModeList {
		t.Error("n should cancel kill")
	}
}

// --- Detach ---

func TestDetach(t *testing.T) {
	m := newTestModel()
	result, cmd := m.Update(key("d"))
	model := result.(*Model)
	if model.Action().Kind != "detach" {
		t.Errorf("want detach, got %v", model.Action())
	}
	if cmd == nil {
		t.Error("should quit after detach")
	}
}

// --- Empty sessions ---

func TestEmptySessions(t *testing.T) {
	m := New(nil)
	_, cmd := m.Update(special(tea.KeyEnter))
	if cmd != nil {
		t.Error("enter with no sessions should not quit")
	}
}
