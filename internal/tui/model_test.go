package tui

import (
	"errors"
	"testing"

	"github.com/blockful/tmux-pilot/internal/tmux"
	tea "github.com/charmbracelet/bubbletea"
)

// helper to create a model pre-loaded with sessions.
func testModel() (*Model, *tmux.MockClient) {
	mock := tmux.NewMockClient()
	m := New(mock)
	// Simulate Init() by processing the sessionsMsg directly
	m.sessions = []tmux.Session{
		{Name: "main", WindowCount: 3, Attached: true},
		{Name: "api-server", WindowCount: 1, Attached: false},
		{Name: "notes", WindowCount: 2, Attached: false},
	}
	return m, mock
}

func key(k string) tea.KeyMsg {
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(k)}
}

func specialKey(t tea.KeyType) tea.KeyMsg {
	return tea.KeyMsg{Type: t}
}

// --- Initialization ---

func TestNew(t *testing.T) {
	mock := tmux.NewMockClient()
	m := New(mock)

	if m.Mode() != ModeList {
		t.Errorf("expected ModeList, got %d", m.Mode())
	}
	if m.Cursor() != 0 {
		t.Errorf("expected cursor 0, got %d", m.Cursor())
	}
	if m.Width() != 80 {
		t.Errorf("expected width 80, got %d", m.Width())
	}
}

func TestInit_ReturnsCommand(t *testing.T) {
	mock := tmux.NewMockClient()
	m := New(mock)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init() should return a command")
	}
}

// --- Window resize ---

func TestWindowResize(t *testing.T) {
	m, _ := testModel()
	result, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	model := result.(*Model)

	if model.Width() != 120 {
		t.Errorf("expected width 120, got %d", model.Width())
	}
}

// --- Sessions message ---

func TestSessionsMsg(t *testing.T) {
	m, _ := testModel()
	sessions := []tmux.Session{
		{Name: "one", WindowCount: 1, Attached: false},
	}

	result, _ := m.Update(sessionsMsg{sessions: sessions})
	model := result.(*Model)

	if len(model.Sessions()) != 1 {
		t.Errorf("expected 1 session, got %d", len(model.Sessions()))
	}
}

func TestSessionsMsg_CursorClamp(t *testing.T) {
	m, _ := testModel()
	m.cursor = 10

	result, _ := m.Update(sessionsMsg{sessions: []tmux.Session{
		{Name: "only", WindowCount: 1, Attached: false},
	}})
	model := result.(*Model)

	if model.Cursor() != 0 {
		t.Errorf("expected cursor clamped to 0, got %d", model.Cursor())
	}
}

func TestSessionsMsg_EmptyList(t *testing.T) {
	m, _ := testModel()
	m.cursor = 2

	result, _ := m.Update(sessionsMsg{sessions: nil})
	model := result.(*Model)

	if model.Cursor() != 0 {
		t.Errorf("expected cursor 0 for empty list, got %d", model.Cursor())
	}
}

func TestSessionsMsg_WithError(t *testing.T) {
	m, _ := testModel()
	testErr := errors.New("connection refused")

	result, _ := m.Update(sessionsMsg{err: testErr})
	model := result.(*Model)

	if model.Err() == nil {
		t.Error("expected error to be set")
	}
}

// --- Navigation ---

func TestNavigation_Down(t *testing.T) {
	m, _ := testModel()
	result, _ := m.Update(key("j"))
	model := result.(*Model)
	if model.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", model.Cursor())
	}
}

func TestNavigation_Up(t *testing.T) {
	m, _ := testModel()
	m.cursor = 2
	result, _ := m.Update(key("k"))
	model := result.(*Model)
	if model.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", model.Cursor())
	}
}

func TestNavigation_DownArrow(t *testing.T) {
	m, _ := testModel()
	result, _ := m.Update(specialKey(tea.KeyDown))
	model := result.(*Model)
	if model.Cursor() != 1 {
		t.Errorf("expected cursor 1, got %d", model.Cursor())
	}
}

func TestNavigation_UpArrow(t *testing.T) {
	m, _ := testModel()
	m.cursor = 1
	result, _ := m.Update(specialKey(tea.KeyUp))
	model := result.(*Model)
	if model.Cursor() != 0 {
		t.Errorf("expected cursor 0, got %d", model.Cursor())
	}
}

func TestNavigation_BoundsTop(t *testing.T) {
	m, _ := testModel()
	m.cursor = 0
	result, _ := m.Update(key("k"))
	model := result.(*Model)
	if model.Cursor() != 0 {
		t.Errorf("cursor should not go below 0, got %d", model.Cursor())
	}
}

func TestNavigation_BoundsBottom(t *testing.T) {
	m, _ := testModel()
	m.cursor = 2 // last index
	result, _ := m.Update(key("j"))
	model := result.(*Model)
	if model.Cursor() != 2 {
		t.Errorf("cursor should not exceed last index, got %d", model.Cursor())
	}
}

// --- Quit ---

func TestQuit_Q(t *testing.T) {
	m, _ := testModel()
	_, cmd := m.Update(key("q"))
	if cmd == nil {
		t.Error("q should produce quit command")
	}
}

func TestQuit_Esc(t *testing.T) {
	m, _ := testModel()
	_, cmd := m.Update(specialKey(tea.KeyEscape))
	if cmd == nil {
		t.Error("esc should produce quit command")
	}
}

// --- Mode transitions ---

func TestEnterCreateMode(t *testing.T) {
	m, _ := testModel()
	result, _ := m.Update(key("n"))
	model := result.(*Model)
	if model.Mode() != ModeCreate {
		t.Errorf("expected ModeCreate, got %d", model.Mode())
	}
	if model.Input() != "" {
		t.Error("input should be empty on entering create mode")
	}
}

func TestEnterRenameMode(t *testing.T) {
	m, _ := testModel()
	m.cursor = 1 // api-server
	result, _ := m.Update(key("r"))
	model := result.(*Model)
	if model.Mode() != ModeRename {
		t.Errorf("expected ModeRename, got %d", model.Mode())
	}
	if model.Input() != "api-server" {
		t.Errorf("expected input pre-filled with 'api-server', got %q", model.Input())
	}
}

func TestEnterRenameMode_EmptySessions(t *testing.T) {
	m, _ := testModel()
	m.sessions = nil
	result, _ := m.Update(key("r"))
	model := result.(*Model)
	if model.Mode() != ModeList {
		t.Error("should stay in list mode with no sessions")
	}
}

func TestEnterConfirmKill(t *testing.T) {
	m, _ := testModel()
	m.cursor = 1
	result, _ := m.Update(key("x"))
	model := result.(*Model)
	if model.Mode() != ModeConfirmKill {
		t.Errorf("expected ModeConfirmKill, got %d", model.Mode())
	}
	if model.KillName() != "api-server" {
		t.Errorf("expected killName 'api-server', got %q", model.KillName())
	}
}

func TestEnterConfirmKill_EmptySessions(t *testing.T) {
	m, _ := testModel()
	m.sessions = nil
	result, _ := m.Update(key("x"))
	model := result.(*Model)
	if model.Mode() != ModeList {
		t.Error("should stay in list mode with no sessions")
	}
}

// --- Input handling ---

func TestInput_TypeCharacters(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeCreate

	for _, ch := range "test" {
		result, _ := m.Update(key(string(ch)))
		m = result.(*Model)
	}

	if m.Input() != "test" {
		t.Errorf("expected 'test', got %q", m.Input())
	}
}

func TestInput_Backspace(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeCreate
	m.input = "test"

	result, _ := m.Update(specialKey(tea.KeyBackspace))
	model := result.(*Model)

	if model.Input() != "tes" {
		t.Errorf("expected 'tes', got %q", model.Input())
	}
}

func TestInput_BackspaceEmpty(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeCreate
	m.input = ""

	result, _ := m.Update(specialKey(tea.KeyBackspace))
	model := result.(*Model)

	if model.Input() != "" {
		t.Error("backspace on empty should stay empty")
	}
}

func TestInput_EscCancels(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeCreate
	m.input = "partial"

	result, _ := m.Update(specialKey(tea.KeyEscape))
	model := result.(*Model)

	if model.Mode() != ModeList {
		t.Error("esc should return to list mode")
	}
	if model.Input() != "" {
		t.Error("input should be cleared on cancel")
	}
}

func TestInput_EnterEmpty(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeCreate
	m.input = ""

	result, cmd := m.Update(specialKey(tea.KeyEnter))
	model := result.(*Model)

	if cmd != nil {
		t.Error("enter with empty input should not produce a command")
	}
	if model.Mode() != ModeCreate {
		t.Error("should stay in create mode with empty input")
	}
}

func TestInput_IgnoresControlChars(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeCreate
	m.input = ""

	// Tab and other control characters should be ignored
	result, _ := m.Update(specialKey(tea.KeyTab))
	model := result.(*Model)

	if model.Input() != "" {
		t.Error("control characters should be ignored")
	}
}

// --- Create mode submit ---

func TestCreate_Submit(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeCreate
	m.input = "new-session"

	_, cmd := m.Update(specialKey(tea.KeyEnter))
	if cmd == nil {
		t.Error("enter with input should produce a command")
	}

	// Execute the command to get the operationMsg
	msg := cmd()
	op, ok := msg.(operationMsg)
	if !ok {
		t.Fatal("expected operationMsg")
	}
	if op.err != nil {
		t.Errorf("unexpected error: %v", op.err)
	}
	if op.switchTo != "new-session" {
		t.Errorf("expected switchTo 'new-session', got %q", op.switchTo)
	}
}

// --- Rename mode submit ---

func TestRename_Submit(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeRename
	m.input = "renamed"
	m.cursor = 0 // "main"

	_, cmd := m.Update(specialKey(tea.KeyEnter))
	if cmd == nil {
		t.Error("enter should produce a command")
	}

	msg := cmd()
	op, ok := msg.(operationMsg)
	if !ok {
		t.Fatal("expected operationMsg")
	}
	if op.err != nil {
		t.Errorf("unexpected error: %v", op.err)
	}
}

// --- Confirm kill ---

func TestConfirmKill_Yes(t *testing.T) {
	m, mock := testModel()
	m.mode = ModeConfirmKill
	m.killName = "notes"

	_, cmd := m.Update(key("y"))
	if cmd == nil {
		t.Error("y should produce kill command")
	}

	msg := cmd()
	op := msg.(operationMsg)
	if op.err != nil {
		t.Errorf("unexpected error: %v", op.err)
	}

	// Verify the session was killed in the mock
	if len(mock.KillCalls) != 1 || mock.KillCalls[0] != "notes" {
		t.Errorf("expected KillCalls=[notes], got %v", mock.KillCalls)
	}
}

func TestConfirmKill_Enter(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeConfirmKill
	m.killName = "notes"

	_, cmd := m.Update(specialKey(tea.KeyEnter))
	if cmd == nil {
		t.Error("enter should confirm kill")
	}
}

func TestConfirmKill_No(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeConfirmKill
	m.killName = "notes"

	result, _ := m.Update(key("n"))
	model := result.(*Model)

	if model.Mode() != ModeList {
		t.Error("n should return to list mode")
	}
	if model.KillName() != "" {
		t.Error("killName should be cleared")
	}
}

func TestConfirmKill_Esc(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeConfirmKill
	m.killName = "notes"

	result, _ := m.Update(specialKey(tea.KeyEscape))
	model := result.(*Model)

	if model.Mode() != ModeList {
		t.Error("esc should return to list mode")
	}
}

// --- Operation message ---

func TestOperationMsg_Success(t *testing.T) {
	m, _ := testModel()
	m.mode = ModeCreate
	m.input = "something"

	result, cmd := m.Update(operationMsg{err: nil})
	model := result.(*Model)

	if model.Mode() != ModeList {
		t.Error("should return to list mode")
	}
	if model.Input() != "" {
		t.Error("input should be cleared")
	}
	if cmd == nil {
		t.Error("should produce refresh command")
	}
}

func TestOperationMsg_WithSwitch(t *testing.T) {
	m, _ := testModel()
	result, cmd := m.Update(operationMsg{switchTo: "new"})
	model := result.(*Model)

	if !model.IsQuitting() {
		t.Error("should be quitting after switch")
	}
	if cmd == nil {
		t.Error("should produce quit command")
	}
}

func TestOperationMsg_Error(t *testing.T) {
	m, _ := testModel()
	testErr := errors.New("failed")

	result, _ := m.Update(operationMsg{err: testErr})
	model := result.(*Model)

	if model.Err() == nil {
		t.Error("error should be set")
	}
	if model.Mode() != ModeList {
		t.Error("should return to list mode even on error")
	}
}

// --- Switch session ---

func TestSwitch_Enter(t *testing.T) {
	m, mock := testModel()
	m.cursor = 1 // api-server

	_, cmd := m.Update(specialKey(tea.KeyEnter))
	if cmd == nil {
		t.Error("enter should produce switch command")
	}

	msg := cmd()
	op := msg.(operationMsg)
	if op.switchTo != "api-server" {
		t.Errorf("expected switchTo 'api-server', got %q", op.switchTo)
	}
	if len(mock.SwitchCalls) != 1 || mock.SwitchCalls[0] != "api-server" {
		t.Errorf("expected SwitchCalls=[api-server], got %v", mock.SwitchCalls)
	}
}

func TestSwitch_EmptySessions(t *testing.T) {
	m, _ := testModel()
	m.sessions = nil

	_, cmd := m.Update(specialKey(tea.KeyEnter))
	if cmd != nil {
		t.Error("enter with no sessions should not produce a command")
	}
}

// --- Error clearing on mode change ---

func TestErrorClearedOnModeChange(t *testing.T) {
	m, _ := testModel()
	m.err = errors.New("stale error")

	result, _ := m.Update(key("n"))
	model := result.(*Model)

	if model.Err() != nil {
		t.Error("error should be cleared when entering create mode")
	}
}
