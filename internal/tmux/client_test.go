package tmux

import (
	"errors"
	"testing"
)

func TestParseSessions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []Session
		wantErr  bool
	}{
		{
			name:  "multiple sessions",
			input: "main\t3\t1\napi-server\t1\t0\nnotes\t2\t0\n",
			expected: []Session{
				{Name: "main", WindowCount: 3, Attached: true},
				{Name: "api-server", WindowCount: 1, Attached: false},
				{Name: "notes", WindowCount: 2, Attached: false},
			},
		},
		{
			name:  "single attached session",
			input: "dev\t5\t1\n",
			expected: []Session{
				{Name: "dev", WindowCount: 5, Attached: true},
			},
		},
		{
			name:     "empty output",
			input:    "",
			expected: nil,
		},
		{
			name:     "whitespace only",
			input:    "  \n  \n",
			expected: nil,
		},
		{
			name:    "invalid window count",
			input:   "main\tnotanumber\t1\n",
			wantErr: true,
		},
		{
			name:  "session name with special characters",
			input: "my-project.v2\t1\t0\n",
			expected: []Session{
				{Name: "my-project.v2", WindowCount: 1, Attached: false},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSessions(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseSessions() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !sessionsEqual(got, tt.expected) {
				t.Errorf("parseSessions() = %+v, want %+v", got, tt.expected)
			}
		})
	}
}

func TestMockClient_ListSessions(t *testing.T) {
	mock := NewMockClient()

	sessions, err := mock.ListSessions()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(sessions) != 3 {
		t.Errorf("expected 3 sessions, got %d", len(sessions))
	}
}

func TestMockClient_ListSessions_Error(t *testing.T) {
	mock := NewMockClient()
	mock.ListErr = errors.New("connection failed")

	_, err := mock.ListSessions()
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestMockClient_NewSession(t *testing.T) {
	mock := NewMockClient()
	initial := len(mock.Sessions())

	if err := mock.NewSession("test"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.Sessions()) != initial+1 {
		t.Errorf("expected %d sessions, got %d", initial+1, len(mock.Sessions()))
	}
	if len(mock.NewCalls) != 1 || mock.NewCalls[0] != "test" {
		t.Errorf("expected NewCalls=[test], got %v", mock.NewCalls)
	}
}

func TestMockClient_NewSession_Error(t *testing.T) {
	mock := NewMockClient()
	mock.NewErr = errors.New("duplicate")

	err := mock.NewSession("main")
	if err == nil {
		t.Error("expected error, got nil")
	}
}

func TestMockClient_SwitchSession(t *testing.T) {
	mock := NewMockClient()

	if err := mock.SwitchSession("api-server"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, s := range mock.Sessions() {
		if s.Name == "api-server" && !s.Attached {
			t.Error("expected api-server to be attached")
		}
		if s.Name == "main" && s.Attached {
			t.Error("expected main to be detached after switch")
		}
	}
}

func TestMockClient_RenameSession(t *testing.T) {
	mock := NewMockClient()

	if err := mock.RenameSession("main", "primary"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	found := false
	for _, s := range mock.Sessions() {
		if s.Name == "primary" {
			found = true
		}
		if s.Name == "main" {
			t.Error("old name should not exist")
		}
	}
	if !found {
		t.Error("renamed session not found")
	}
}

func TestMockClient_RenameSession_NotFound(t *testing.T) {
	mock := NewMockClient()

	err := mock.RenameSession("nonexistent", "new")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestMockClient_KillSession(t *testing.T) {
	mock := NewMockClient()
	initial := len(mock.Sessions())

	if err := mock.KillSession("main"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(mock.Sessions()) != initial-1 {
		t.Errorf("expected %d sessions, got %d", initial-1, len(mock.Sessions()))
	}
	for _, s := range mock.Sessions() {
		if s.Name == "main" {
			t.Error("session should have been killed")
		}
	}
}

func TestMockClient_KillSession_NotFound(t *testing.T) {
	mock := NewMockClient()

	err := mock.KillSession("ghost")
	if err == nil {
		t.Error("expected error for nonexistent session")
	}
}

func TestMockClient_IsInsideTmux(t *testing.T) {
	mock := NewMockClient()

	if !mock.IsInsideTmux() {
		t.Error("expected true by default")
	}

	mock.SetInsideTmux(false)
	if mock.IsInsideTmux() {
		t.Error("expected false after SetInsideTmux(false)")
	}
}

// sessionsEqual compares two session slices.
func sessionsEqual(a, b []Session) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
