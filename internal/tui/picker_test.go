package tui

import (
	"testing"

	"github.com/blockful/tmux-pilot/internal/tmux"
)

func TestState_Navigation(t *testing.T) {
	sessions := []tmux.Session{
		{Name: "session1", WindowCount: 1, Attached: false},
		{Name: "session2", WindowCount: 1, Attached: false},
		{Name: "session3", WindowCount: 1, Attached: false},
	}

	state := &State{
		sessions: sessions,
		cursor:   0,
		mode:     ModeList,
	}

	// Test down navigation
	err := state.handleKey(Key{Type: KeyDown})
	if err != nil {
		t.Fatal(err)
	}
	if state.cursor != 1 {
		t.Errorf("Expected cursor 1, got %d", state.cursor)
	}

	// Test up navigation
	err = state.handleKey(Key{Type: KeyUp})
	if err != nil {
		t.Fatal(err)
	}
	if state.cursor != 0 {
		t.Errorf("Expected cursor 0, got %d", state.cursor)
	}

	// Test vi-style navigation
	err = state.handleKey(Key{Type: KeyRune, Rune: 'j'})
	if err != nil {
		t.Fatal(err)
	}
	if state.cursor != 1 {
		t.Errorf("Expected cursor 1, got %d", state.cursor)
	}

	err = state.handleKey(Key{Type: KeyRune, Rune: 'k'})
	if err != nil {
		t.Fatal(err)
	}
	if state.cursor != 0 {
		t.Errorf("Expected cursor 0, got %d", state.cursor)
	}
}

func TestState_BoundaryChecks(t *testing.T) {
	sessions := []tmux.Session{
		{Name: "session1", WindowCount: 1, Attached: false},
	}

	state := &State{
		sessions: sessions,
		cursor:   0,
		mode:     ModeList,
	}

	// Test up at boundary
	err := state.handleKey(Key{Type: KeyUp})
	if err != nil {
		t.Fatal(err)
	}
	if state.cursor != 0 {
		t.Errorf("Cursor should stay at 0, got %d", state.cursor)
	}

	// Test down at boundary  
	err = state.handleKey(Key{Type: KeyDown})
	if err != nil {
		t.Fatal(err)
	}
	if state.cursor != 0 {
		t.Errorf("Cursor should stay at 0, got %d", state.cursor)
	}
}

func TestState_ModeTransitions(t *testing.T) {
	sessions := []tmux.Session{
		{Name: "session1", WindowCount: 1, Attached: false},
	}

	state := &State{
		sessions: sessions,
		cursor:   0,
		mode:     ModeList,
	}

	// Test create mode transition
	err := state.handleKey(Key{Type: KeyRune, Rune: 'n'})
	if err != nil {
		t.Fatal(err)
	}
	if state.mode != ModeCreate {
		t.Errorf("Expected ModeCreate, got %v", state.mode)
	}
	if state.input != "" {
		t.Errorf("Expected empty input, got %q", state.input)
	}

	// Test escape back to list
	err = state.handleKey(Key{Type: KeyEscape})
	if err != nil {
		t.Fatal(err)
	}
	if state.mode != ModeList {
		t.Errorf("Expected ModeList, got %v", state.mode)
	}

	// Test rename mode transition
	err = state.handleKey(Key{Type: KeyRune, Rune: 'r'})
	if err != nil {
		t.Fatal(err)
	}
	if state.mode != ModeRename {
		t.Errorf("Expected ModeRename, got %v", state.mode)
	}
	if state.input != "session1" {
		t.Errorf("Expected input 'session1', got %q", state.input)
	}

	// Test kill confirm mode
	err = state.handleKey(Key{Type: KeyEscape})
	if err != nil {
		t.Fatal(err)
	}
	err = state.handleKey(Key{Type: KeyRune, Rune: 'x'})
	if err != nil {
		t.Fatal(err)
	}
	if state.mode != ModeConfirmKill {
		t.Errorf("Expected ModeConfirmKill, got %v", state.mode)
	}
}

func TestState_Actions(t *testing.T) {
	sessions := []tmux.Session{
		{Name: "session1", WindowCount: 1, Attached: false},
	}

	tests := []struct {
		name           string
		setupMode      Mode
		input          string
		key            Key
		expectedAction Action
		expectedDone   bool
	}{
		{
			"switch action",
			ModeList,
			"",
			Key{Type: KeyEnter},
			Action{Kind: "switch", Target: "session1"},
			true,
		},
		{
			"detach action",
			ModeList,
			"",
			Key{Type: KeyRune, Rune: 'd'},
			Action{Kind: "detach"},
			true,
		},
		{
			"quit action",
			ModeList,
			"",
			Key{Type: KeyRune, Rune: 'q'},
			Action{},
			true,
		},
		{
			"new session action",
			ModeCreate,
			"newsession",
			Key{Type: KeyEnter},
			Action{Kind: "new", Target: "newsession"},
			true,
		},
		{
			"rename action",
			ModeRename,
			"renamed",
			Key{Type: KeyEnter},
			Action{Kind: "rename", Target: "session1", NewName: "renamed"},
			true,
		},
		{
			"kill action",
			ModeConfirmKill,
			"",
			Key{Type: KeyRune, Rune: 'y'},
			Action{Kind: "kill", Target: "session1"},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := &State{
				sessions: sessions,
				cursor:   0,
				mode:     tt.setupMode,
				input:    tt.input,
			}

			err := state.handleKey(tt.key)
			if err != nil {
				t.Fatal(err)
			}

			if state.done != tt.expectedDone {
				t.Errorf("Expected done=%v, got %v", tt.expectedDone, state.done)
			}

			if state.action != tt.expectedAction {
				t.Errorf("Expected action %+v, got %+v", tt.expectedAction, state.action)
			}
		})
	}
}

func TestState_InputHandling(t *testing.T) {
	state := &State{
		mode:  ModeCreate,
		input: "",
	}

	// Test character input
	err := state.handleKey(Key{Type: KeyRune, Rune: 'a'})
	if err != nil {
		t.Fatal(err)
	}
	if state.input != "a" {
		t.Errorf("Expected input 'a', got %q", state.input)
	}

	// Test backspace
	err = state.handleKey(Key{Type: KeyBackspace})
	if err != nil {
		t.Fatal(err)
	}
	if state.input != "" {
		t.Errorf("Expected empty input, got %q", state.input)
	}

	// Test multiple characters
	for _, ch := range "test123" {
		err = state.handleKey(Key{Type: KeyRune, Rune: ch})
		if err != nil {
			t.Fatal(err)
		}
	}
	if state.input != "test123" {
		t.Errorf("Expected input 'test123', got %q", state.input)
	}
}

func TestState_EmptySessionsList(t *testing.T) {
	state := &State{
		sessions: []tmux.Session{},
		cursor:   0,
		mode:     ModeList,
	}

	// Navigation should not crash on empty list
	err := state.handleKey(Key{Type: KeyDown})
	if err != nil {
		t.Fatal(err)
	}
	err = state.handleKey(Key{Type: KeyUp})
	if err != nil {
		t.Fatal(err)
	}

	// Enter should not trigger action on empty list
	err = state.handleKey(Key{Type: KeyEnter})
	if err != nil {
		t.Fatal(err)
	}
	if state.done {
		t.Error("Should not complete action with empty session list")
	}

	// Rename/kill should not be available
	err = state.handleKey(Key{Type: KeyRune, Rune: 'r'})
	if err != nil {
		t.Fatal(err)
	}
	if state.mode != ModeList {
		t.Error("Rename should not be available for empty list")
	}

	err = state.handleKey(Key{Type: KeyRune, Rune: 'x'})
	if err != nil {
		t.Fatal(err)
	}
	if state.mode != ModeList {
		t.Error("Kill should not be available for empty list")
	}
}