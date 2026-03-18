package tui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/blockful/tmux-pilot/internal/tmux"
)

// isolatedTmux creates a temporary tmux server with an isolated socket.
// Returns ClientOptions and a cleanup function.
func isolatedTmux(t *testing.T) (tmux.ClientOptions, func()) {
	t.Helper()
	sock := filepath.Join(t.TempDir(), "tmux-test.sock")
	opts := tmux.ClientOptions{SocketPath: sock}

	// Start a detached session so the server exists
	args := append(opts.Args(), "new-session", "-d", "-s", "init")
	cmd := exec.Command("tmux", args...)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to start isolated tmux: %v", err)
	}

	cleanup := func() {
		killArgs := append(opts.Args(), "kill-server")
		_ = exec.Command("tmux", killArgs...).Run()
		_ = os.Remove(sock)
	}

	return opts, cleanup
}

// addTestSession creates a session in the isolated tmux server.
func addTestSession(t *testing.T, name string, opts tmux.ClientOptions) {
	t.Helper()
	args := append(opts.Args(), "new-session", "-d", "-s", name)
	if err := exec.Command("tmux", args...).Run(); err != nil {
		t.Fatalf("failed to create test session %q: %v", name, err)
	}
}

// listTestSessions returns sessions from the isolated server.
func listTestSessions(t *testing.T, opts tmux.ClientOptions) []tmux.Session {
	t.Helper()
	sessions, err := tmux.ListSessions(opts)
	if err != nil {
		t.Fatalf("failed to list sessions: %v", err)
	}
	return sessions
}

// hasSession checks if a named session exists in a list.
func hasSession(sessions []tmux.Session, name string) bool {
	for _, s := range sessions {
		if s.Name == name {
			return true
		}
	}
	return false
}

var _ = fmt.Sprintf // ensure fmt is used

func TestState_Navigation(t *testing.T) {
	sessions := []tmux.Session{
		{Name: "session1", WindowCount: 1, Attached: false},
		{Name: "session2", WindowCount: 1, Attached: false},
		{Name: "session3", WindowCount: 1, Attached: false},
	}

	state := &State{sessions: sessions, cursor: 0, mode: ModeList}

	// Down
	_ = state.handleKey(Key{Type: KeyDown})
	if state.cursor != 1 {
		t.Errorf("Expected cursor 1, got %d", state.cursor)
	}

	// Up
	_ = state.handleKey(Key{Type: KeyUp})
	if state.cursor != 0 {
		t.Errorf("Expected cursor 0, got %d", state.cursor)
	}

	// j/k
	_ = state.handleKey(Key{Type: KeyRune, Rune: 'j'})
	if state.cursor != 1 {
		t.Errorf("Expected cursor 1 after j, got %d", state.cursor)
	}
	_ = state.handleKey(Key{Type: KeyRune, Rune: 'k'})
	if state.cursor != 0 {
		t.Errorf("Expected cursor 0 after k, got %d", state.cursor)
	}
}

func TestState_BoundaryChecks(t *testing.T) {
	state := &State{
		sessions: []tmux.Session{{Name: "only", WindowCount: 1}},
		cursor:   0,
		mode:     ModeList,
	}

	_ = state.handleKey(Key{Type: KeyUp})
	if state.cursor != 0 {
		t.Errorf("Cursor should stay at 0, got %d", state.cursor)
	}
	_ = state.handleKey(Key{Type: KeyDown})
	if state.cursor != 0 {
		t.Errorf("Cursor should stay at 0, got %d", state.cursor)
	}
}

func TestState_ModeTransitions(t *testing.T) {
	sessions := []tmux.Session{{Name: "session1", WindowCount: 1}}
	state := &State{sessions: sessions, cursor: 0, mode: ModeList}

	// n → ModeCreate
	_ = state.handleKey(Key{Type: KeyRune, Rune: 'n'})
	if state.mode != ModeCreate {
		t.Errorf("Expected ModeCreate, got %v", state.mode)
	}
	if state.input != "" {
		t.Errorf("Expected empty input, got %q", state.input)
	}

	// Esc → ModeList
	_ = state.handleKey(Key{Type: KeyEscape})
	if state.mode != ModeList {
		t.Errorf("Expected ModeList, got %v", state.mode)
	}

	// r → ModeRename (pre-fills name)
	_ = state.handleKey(Key{Type: KeyRune, Rune: 'r'})
	if state.mode != ModeRename {
		t.Errorf("Expected ModeRename, got %v", state.mode)
	}
	if state.input != "session1" {
		t.Errorf("Expected input 'session1', got %q", state.input)
	}

	// Esc → x → ModeConfirmKill
	_ = state.handleKey(Key{Type: KeyEscape})
	_ = state.handleKey(Key{Type: KeyRune, Rune: 'x'})
	if state.mode != ModeConfirmKill {
		t.Errorf("Expected ModeConfirmKill, got %v", state.mode)
	}
}

func TestState_ExitActions(t *testing.T) {
	sessions := []tmux.Session{{Name: "session1", WindowCount: 1}}

	tests := []struct {
		name           string
		key            Key
		expectedAction Action
		expectedDone   bool
	}{
		{
			"switch on enter",
			Key{Type: KeyEnter},
			Action{Kind: "switch", Target: "session1"},
			true,
		},

		{
			"quit on q",
			Key{Type: KeyRune, Rune: 'q'},
			Action{},
			true,
		},
		{
			"quit on escape",
			Key{Type: KeyEscape},
			Action{},
			true,
		},
		{
			"quit on ctrl-c",
			Key{Type: KeyCtrlC},
			Action{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &State{sessions: sessions, cursor: 0, mode: ModeList}
			_ = state.handleKey(tt.key)

			if state.done != tt.expectedDone {
				t.Errorf("Expected done=%v, got %v", tt.expectedDone, state.done)
			}
			if state.action != tt.expectedAction {
				t.Errorf("Expected action %+v, got %+v", tt.expectedAction, state.action)
			}
		})
	}
}

func TestState_KillStaysOpen(t *testing.T) {
	opts, cleanup := isolatedTmux(t)
	defer cleanup()

	addTestSession(t, "tokill", opts)
	sessions := listTestSessions(t, opts)

	// Find the "tokill" session index
	killIdx := -1
	for i, s := range sessions {
		if s.Name == "tokill" {
			killIdx = i
			break
		}
	}
	if killIdx == -1 {
		t.Fatal("tokill session not found")
	}

	state := &State{
		sessions:   sessions,
		cursor:     killIdx,
		mode:       ModeConfirmKill,
		clientOpts: opts,
	}

	_ = state.handleConfirmKey(Key{Type: KeyRune, Rune: 'y'})

	if state.done {
		t.Error("Kill should NOT exit the picker")
	}
	if state.mode != ModeList {
		t.Errorf("Expected ModeList after kill, got %v", state.mode)
	}
	// Verify session was actually killed
	if hasSession(state.sessions, "tokill") {
		t.Error("Session 'tokill' should have been removed")
	}
}

func TestState_KillConfirmNo(t *testing.T) {
	sessions := []tmux.Session{{Name: "s1", WindowCount: 1}}
	state := &State{sessions: sessions, cursor: 0, mode: ModeConfirmKill}

	_ = state.handleConfirmKey(Key{Type: KeyRune, Rune: 'n'})

	if state.done {
		t.Error("Cancel should not exit")
	}
	if state.mode != ModeList {
		t.Errorf("Expected ModeList, got %v", state.mode)
	}
}

func TestState_InputHandling(t *testing.T) {
	state := &State{mode: ModeCreate, input: ""}

	// Type characters
	_ = state.handleKey(Key{Type: KeyRune, Rune: 'a'})
	if state.input != "a" {
		t.Errorf("Expected 'a', got %q", state.input)
	}

	// Backspace
	_ = state.handleKey(Key{Type: KeyBackspace})
	if state.input != "" {
		t.Errorf("Expected empty, got %q", state.input)
	}

	// Type multiple
	for _, ch := range "test123" {
		_ = state.handleKey(Key{Type: KeyRune, Rune: ch})
	}
	if state.input != "test123" {
		t.Errorf("Expected 'test123', got %q", state.input)
	}
}

func TestState_EmptySessionsList(t *testing.T) {
	state := &State{sessions: []tmux.Session{}, cursor: 0, mode: ModeList}

	// Navigation on empty list should not crash
	_ = state.handleKey(Key{Type: KeyDown})
	_ = state.handleKey(Key{Type: KeyUp})

	// Enter should not trigger action
	_ = state.handleKey(Key{Type: KeyEnter})
	if state.done {
		t.Error("Should not complete with empty session list")
	}

	// r/x should not change mode on empty list
	_ = state.handleKey(Key{Type: KeyRune, Rune: 'r'})
	if state.mode != ModeList {
		t.Error("Rename should not activate on empty list")
	}
	_ = state.handleKey(Key{Type: KeyRune, Rune: 'x'})
	if state.mode != ModeList {
		t.Error("Kill should not activate on empty list")
	}
}

func TestState_CreateStaysOpen(t *testing.T) {
	opts, cleanup := isolatedTmux(t)
	defer cleanup()

	sessions := listTestSessions(t, opts)

	state := &State{
		sessions:   sessions,
		cursor:     0,
		mode:       ModeCreate,
		input:      "brandnew",
		clientOpts: opts,
	}

	_ = state.handleInputKey(Key{Type: KeyEnter}, "new")

	if state.done {
		t.Error("Create should NOT exit the picker")
	}
	if state.mode != ModeList {
		t.Errorf("Expected ModeList after create, got %v", state.mode)
	}
	// Verify session was actually created
	if !hasSession(state.sessions, "brandnew") {
		t.Error("Session 'brandnew' should exist after creation")
	}
}

func TestState_RenameStaysOpen(t *testing.T) {
	opts, cleanup := isolatedTmux(t)
	defer cleanup()

	addTestSession(t, "oldname", opts)
	sessions := listTestSessions(t, opts)

	// Find oldname index
	idx := -1
	for i, s := range sessions {
		if s.Name == "oldname" {
			idx = i
			break
		}
	}
	if idx == -1 {
		t.Fatal("oldname session not found")
	}

	state := &State{
		sessions:   sessions,
		cursor:     idx,
		mode:       ModeRename,
		input:      "newname",
		clientOpts: opts,
	}

	_ = state.handleInputKey(Key{Type: KeyEnter}, "rename")

	if state.done {
		t.Error("Rename should NOT exit the picker")
	}
	if state.mode != ModeList {
		t.Errorf("Expected ModeList after rename, got %v", state.mode)
	}
	// Verify rename actually happened
	if hasSession(state.sessions, "oldname") {
		t.Error("Old session name should be gone")
	}
	if !hasSession(state.sessions, "newname") {
		t.Error("New session name should exist")
	}
}
