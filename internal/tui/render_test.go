package tui

import (
	"bytes"
	"strings"
	"testing"

	"github.com/blockful/tmux-pilot/internal/tmux"
)

func TestRenderer_ANSICodes(t *testing.T) {
	tests := []struct {
		name     string
		function func(*Renderer)
		expected string
	}{
		{"MoveCursor", func(r *Renderer) { r.MoveCursor(10, 5) }, "\x1b[5;10H"},
		{"MoveUp", func(r *Renderer) { r.MoveUp(3) }, "\x1b[3A"},
		{"ClearLine", func(r *Renderer) { r.ClearLine() }, "\x1b[2K"},
		{"ClearFromCursor", func(r *Renderer) { r.ClearFromCursor() }, "\x1b[0K"},
		{"HideCursor", func(r *Renderer) { r.HideCursor() }, "\x1b[?25l"},
		{"ShowCursor", func(r *Renderer) { r.ShowCursor() }, "\x1b[?25h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			renderer := NewRendererTo(&buf, ColorEnabled)
			tt.function(renderer)

			if got := buf.String(); got != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, got)
			}
		})
	}
}

func TestRenderer_MoveUpZero(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererTo(&buf, ColorEnabled)
	r.MoveUp(0)
	if buf.Len() != 0 {
		t.Errorf("MoveUp(0) should be a no-op, got %q", buf.String())
	}
}

func TestRenderer_ColorModes(t *testing.T) {
	tests := []struct {
		name        string
		colorMode   ColorMode
		expectColor bool
	}{
		{"colors enabled", ColorEnabled, true},
		{"colors disabled", ColorDisabled, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := NewRendererTo(&bytes.Buffer{}, tt.colorMode)
			c := r.color(ColorAccent)
			hasCode := strings.Contains(c, "\x1b[38;5;")

			if tt.expectColor && !hasCode {
				t.Error("Expected color code but got none")
			}
			if !tt.expectColor && hasCode {
				t.Error("Expected no color code but got one")
			}
		})
	}
}

func TestRenderer_RenderUI(t *testing.T) {
	sessions := []tmux.Session{
		{Name: "main", WindowCount: 3, Attached: true},
		{Name: "api", WindowCount: 1, Attached: false},
	}

	tests := []struct {
		name     string
		mode     Mode
		input    string
		warning  string
		contains []string
	}{
		{
			"list mode",
			ModeList,
			"", "",
			[]string{"tmux sessions", "main", "api", "navigate"},
		},
		{
			"create mode",
			ModeCreate,
			"new-session", "",
			[]string{"New session name:", "new-session", "confirm"},
		},
		{
			"rename mode with warning",
			ModeRename,
			"duplicate", "already exists",
			[]string{"Rename to:", "duplicate", "already exists"},
		},
		{
			"confirm kill",
			ModeConfirmKill,
			"", "",
			[]string{"Kill session", "main"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			renderer := NewRendererTo(&buf, ColorDisabled)
			renderer.RenderUI(sessions, 0, tt.mode, tt.input, tt.warning, 80)

			output := buf.String()
			for _, expected := range tt.contains {
				if !strings.Contains(output, expected) {
					t.Errorf("Output should contain %q, got:\n%s", expected, output)
				}
			}
		})
	}
}

func TestRenderer_RenderUI_RewindsOnSecondCall(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererTo(&buf, ColorDisabled)
	sessions := []tmux.Session{{Name: "s1", WindowCount: 1, Attached: false}}

	r.RenderUI(sessions, 0, ModeList, "", "", 60)
	first := buf.String()

	r.RenderUI(sessions, 0, ModeList, "", "", 60)
	second := buf.String()[len(first):]

	// Second render should start with a cursor-up sequence to rewind
	if !strings.Contains(second, "\x1b[") || !strings.Contains(second, "A") {
		t.Error("Second render should rewind cursor with MoveUp escape")
	}
}

func TestRenderer_EmptySessionsList(t *testing.T) {
	var buf bytes.Buffer
	renderer := NewRendererTo(&buf, ColorDisabled)
	renderer.RenderUI([]tmux.Session{}, 0, ModeList, "", "", 80)

	output := buf.String()
	if !strings.Contains(output, "No tmux sessions") {
		t.Error("Should display 'No tmux sessions' message for empty list")
	}
}

func TestRenderer_Cleanup(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererTo(&buf, ColorDisabled)
	sessions := []tmux.Session{{Name: "s1", WindowCount: 1, Attached: false}}

	r.RenderUI(sessions, 0, ModeList, "", "", 60)
	buf.Reset()

	r.Cleanup()
	output := buf.String()

	// Cleanup should contain cursor movement and clear sequences
	if !strings.Contains(output, "\x1b[") {
		t.Error("Cleanup should emit ANSI sequences")
	}
	// Should show cursor
	if !strings.Contains(output, "\x1b[?25h") {
		t.Error("Cleanup should show cursor")
	}
}

func TestRenderer_UsesRawNewlines(t *testing.T) {
	var buf bytes.Buffer
	r := NewRendererTo(&buf, ColorDisabled)
	sessions := []tmux.Session{{Name: "test", WindowCount: 1, Attached: false}}

	r.RenderUI(sessions, 0, ModeList, "", "", 60)
	output := buf.String()

	// In raw mode we need \r\n. The writeln method prepends \r to each line.
	if !strings.Contains(output, "\r\n") {
		t.Error("Output should use \\r\\n for raw mode compatibility")
	}
}
