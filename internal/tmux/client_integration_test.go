//go:build integration

package tmux

import (
	"os/exec"
	"testing"
	"time"
)

func isTmuxAvailable() bool {
	_, err := exec.LookPath("tmux")
	return err == nil
}

func TestRealClient_Integration(t *testing.T) {
	if !isTmuxAvailable() {
		t.Skip("tmux not available")
	}

	client := NewRealClient()
	testSession := "tmux-pilot-integration-test"

	// Cleanup before and after
	_ = client.KillSession(testSession)
	defer func() { _ = client.KillSession(testSession) }()

	// Create
	if err := client.NewSession(testSession); err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	time.Sleep(100 * time.Millisecond)

	// List and find
	sessions, err := client.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}

	found := false
	for _, s := range sessions {
		if s.Name == testSession {
			found = true
			if s.WindowCount < 1 {
				t.Errorf("expected at least 1 window, got %d", s.WindowCount)
			}
		}
	}
	if !found {
		t.Fatalf("session %q not found", testSession)
	}

	// Rename
	renamed := testSession + "-renamed"
	if err := client.RenameSession(testSession, renamed); err != nil {
		t.Fatalf("RenameSession: %v", err)
	}
	defer func() { _ = client.KillSession(renamed) }()

	sessions, err = client.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions after rename: %v", err)
	}
	found = false
	for _, s := range sessions {
		if s.Name == renamed {
			found = true
		}
		if s.Name == testSession {
			t.Error("old session name still exists")
		}
	}
	if !found {
		t.Fatalf("renamed session %q not found", renamed)
	}

	// Kill
	if err := client.KillSession(renamed); err != nil {
		t.Fatalf("KillSession: %v", err)
	}

	sessions, err = client.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions after kill: %v", err)
	}
	for _, s := range sessions {
		if s.Name == renamed {
			t.Errorf("session %q still exists after kill", renamed)
		}
	}
}

func TestRealClient_IsInsideTmux(t *testing.T) {
	if !isTmuxAvailable() {
		t.Skip("tmux not available")
	}
	client := NewRealClient()
	// Just ensure it doesn't panic; result depends on test environment
	_ = client.IsInsideTmux()
}
