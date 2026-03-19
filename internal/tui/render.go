package tui

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/blockful/tmux-pilot/internal/tmux"
)

// ColorMode controls ANSI color output.
type ColorMode bool

const (
	ColorEnabled  ColorMode = true
	ColorDisabled ColorMode = false
)

// ANSI color codes (256-color).
const (
	ColorAccent = 205 // bright magenta
	ColorText   = 252 // light gray
	ColorDim    = 243 // dark gray
	ColorBorder = 240 // darker gray
	ColorGreen  = 46  // bright green
	ColorYellow = 214 // orange-yellow
	ColorBg     = 237 // dark background
)

// Renderer handles ANSI output and screen drawing.
// All output goes through the writer (stdout by default) and uses \r\n
// because the terminal is in raw mode (no automatic CR on LF).
type Renderer struct {
	colorMode  ColorMode
	w          io.Writer
	lastHeight int // lines rendered last frame, for cursor rewind
}

// NewRenderer creates a new renderer writing to stdout.
func NewRenderer(colorMode ColorMode) *Renderer {
	return &Renderer{colorMode: colorMode, w: os.Stdout}
}

// NewRendererTo creates a renderer writing to a custom writer (for testing).
func NewRendererTo(w io.Writer, colorMode ColorMode) *Renderer {
	return &Renderer{colorMode: colorMode, w: w}
}

// --- ANSI primitives ---
//
// Terminal escape writes intentionally discard errors — if stdout is broken
// mid-render, there is no recovery path. We use a helper to satisfy linters.

// write is a helper that writes to the output and discards errors.
func (r *Renderer) write(s string) {
	_, _ = fmt.Fprint(r.w, s)
}

// writef is a formatted write helper that discards errors.
func (r *Renderer) writef(format string, args ...any) {
	_, _ = fmt.Fprintf(r.w, format, args...)
}

// MoveCursor positions the cursor at (x, y) coordinates (1-indexed).
func (r *Renderer) MoveCursor(x, y int) {
	r.writef("\x1b[%d;%dH", y, x)
}

// MoveUp moves the cursor up n lines.
func (r *Renderer) MoveUp(n int) {
	if n > 0 {
		r.writef("\x1b[%dA", n)
	}
}

// ClearLine clears the entire current line.
func (r *Renderer) ClearLine() {
	r.write("\x1b[2K")
}

// ClearFromCursor clears from cursor to end of line.
func (r *Renderer) ClearFromCursor() {
	r.write("\x1b[0K")
}

// HideCursor hides the cursor.
func (r *Renderer) HideCursor() {
	r.write("\x1b[?25l")
}

// ShowCursor shows the cursor.
func (r *Renderer) ShowCursor() {
	r.write("\x1b[?25h")
}

// color writes a foreground color escape if colors are enabled.
func (r *Renderer) color(code int) string {
	if r.colorMode {
		return fmt.Sprintf("\x1b[38;5;%dm", code)
	}
	return ""
}

// bg writes a background color escape if colors are enabled.
func (r *Renderer) bg(code int) string {
	if r.colorMode {
		return fmt.Sprintf("\x1b[48;5;%dm", code)
	}
	return ""
}

// reset returns the ANSI reset sequence.
func (r *Renderer) reset() string {
	return "\x1b[0m"
}

// bold returns the ANSI bold sequence.
func (r *Renderer) bold() string {
	if r.colorMode {
		return "\x1b[1m"
	}
	return ""
}

// --- line helper ---

// writeln writes a line with \r\n (raw mode needs explicit CR).
// It clears the line first to avoid artifacts from previous renders.
func (r *Renderer) writeln(s string) {
	r.writef("\x1b[2K\r%s\r\n", s)
}

// --- High-level rendering ---

// RenderUI draws the complete interface. On subsequent calls it rewinds
// the cursor to overwrite the previous frame, giving in-place updates.
func (r *Renderer) RenderUI(sessions []tmux.Session, cursor int, mode Mode, input, warning string, width int) {
	r.HideCursor()

	// Rewind cursor to top of previous frame
	if r.lastHeight > 0 {
		r.MoveUp(r.lastHeight)
		r.write("\r") // back to column 1
	}

	lines := 0

	// === Session list with border ===
	bdr := r.color(ColorBorder)
	rst := r.reset()

	// Top border
	r.writeln(bdr + "╭" + strings.Repeat("─", width-2) + "╮" + rst)
	lines++

	// Title
	title := "tmux sessions"
	pad := (width - 2 - len(title)) / 2
	if pad < 1 {
		pad = 1
	}
	r.writeln(bdr + "│" + rst +
		strings.Repeat(" ", pad) +
		r.color(ColorAccent) + r.bold() + title + rst +
		strings.Repeat(" ", width-2-pad-len(title)) +
		bdr + "│" + rst)
	lines++

	// Empty line
	r.writeln(bdr + "│" + strings.Repeat(" ", width-2) + "│" + rst)
	lines++

	// Sessions or "no sessions"
	if len(sessions) == 0 {
		msg := "No tmux sessions running"
		p := (width - 2 - len(msg)) / 2
		if p < 1 {
			p = 1
		}
		r.writeln(bdr + "│" + rst +
			strings.Repeat(" ", p) +
			r.color(ColorDim) + msg + rst +
			strings.Repeat(" ", width-2-p-len(msg)) +
			bdr + "│" + rst)
		lines++
	} else {
		// Find longest session name for column alignment
		maxName := 0
		for _, s := range sessions {
			if len(s.Name) > maxName {
				maxName = len(s.Name)
			}
		}
		if maxName < 8 {
			maxName = 8 // minimum column width
		}

		for i, s := range sessions {
			r.writeln(r.fmtSession(s, i == cursor, width, maxName))
			lines++
		}
	}

	// Empty line
	r.writeln(bdr + "│" + strings.Repeat(" ", width-2) + "│" + rst)
	lines++

	// Bottom border
	r.writeln(bdr + "╰" + strings.Repeat("─", width-2) + "╯" + rst)
	lines++

	// === Mode-specific content below the box ===
	switch mode {
	case ModeList:
		r.writeln(r.color(ColorDim) + "↑/k ↓/j: navigate  Enter: attach  n: new  r: rename  x: kill  q/Esc: quit" + rst)
		lines++
		r.writeln(r.color(ColorDim) + "tip: " + rst + r.color(ColorAccent) + "Ctrl-b d" + rst + r.color(ColorDim) + " to detach from tmux" + rst)
		lines++
	case ModeCreate:
		r.writeln(r.color(ColorAccent) + "New session name: " + rst + input + "█")
		lines++
		if warning != "" {
			r.writeln(r.color(ColorYellow) + "⚠ " + warning + rst)
			lines++
		}
		r.writeln(r.color(ColorDim) + "Enter: confirm  Esc: cancel" + rst)
		lines++
	case ModeRename:
		label := "Rename to: "
		r.writeln(r.color(ColorAccent) + label + rst + input + "█")
		lines++
		if warning != "" {
			r.writeln(r.color(ColorYellow) + "⚠ " + warning + rst)
			lines++
		}
		r.writeln(r.color(ColorDim) + "Enter: confirm  Esc: cancel" + rst)
		lines++
	case ModeConfirmKill:
		if len(sessions) > 0 {
			r.writeln(r.color(ColorYellow) + fmt.Sprintf("Kill session '%s'? ", sessions[cursor].Name) + rst + "(y/N)")
			lines++
		}
	}

	// Clear any leftover lines from a taller previous frame
	if lines < r.lastHeight {
		for i := 0; i < r.lastHeight-lines; i++ {
			r.writeln("")
		}
		// Move back up past the blank lines so next rewind is correct
		r.MoveUp(r.lastHeight - lines)
	}

	r.lastHeight = lines
}

// Cleanup clears the rendered UI from the terminal. Call before exiting.
func (r *Renderer) Cleanup() {
	if r.lastHeight > 0 {
		r.MoveUp(r.lastHeight)
		r.write("\r")
		for i := 0; i < r.lastHeight; i++ {
			r.write("\x1b[2K\r\n")
		}
		r.MoveUp(r.lastHeight)
		r.write("\r")
	}
	r.ShowCursor()
}

// fmtSession formats a single session line inside the border box.
// nameWidth is the column width for session names (all rows use the same).
func (r *Renderer) fmtSession(s tmux.Session, selected bool, width int, nameWidth int) string {
	bdr := r.color(ColorBorder)
	rst := r.reset()

	status := "detached"
	if s.Attached {
		status = "attached"
	}
	winWord := "window"
	if s.WindowCount != 1 {
		winWord = "windows"
	}
	// Aligned columns: " ● name     N windows   status"
	infoText := fmt.Sprintf("%-*s  %d %-7s  %s", nameWidth, s.Name, s.WindowCount, winWord, status)
	displayWidth := 1 + 1 + 1 + len(infoText) // space + dot(1col) + space + info

	// Pad to fill the box width (subtract 2 for borders)
	innerWidth := width - 2
	padLen := innerWidth - displayWidth
	if padLen < 0 {
		padLen = 0
	}

	// Build the styled line
	var line strings.Builder
	line.WriteString(bdr + "│" + rst)

	if selected {
		line.WriteString(r.bg(ColorBg))
	}

	// Dot color
	if s.Attached {
		line.WriteString(r.color(ColorGreen))
		line.WriteString(" ● ")
	} else {
		line.WriteString(r.color(ColorDim))
		line.WriteString(" ○ ")
	}

	// Session info
	if selected {
		line.WriteString(r.color(ColorAccent))
	} else {
		line.WriteString(r.color(ColorText))
	}
	line.WriteString(infoText)

	// Padding (keep bg color for selection highlight)
	line.WriteString(strings.Repeat(" ", padLen))
	line.WriteString(rst)
	line.WriteString(bdr + "│" + rst)

	return line.String()
}
