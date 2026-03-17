package tui

import (
	"fmt"
	"strings"

	"github.com/blockful/tmux-pilot/internal/tmux"
	"github.com/charmbracelet/lipgloss"
)

// --- Color palette ---

var (
	accent   = lipgloss.Color("205") // magenta/pink
	text     = lipgloss.Color("252") // light gray
	dim      = lipgloss.Color("243") // muted gray
	border   = lipgloss.Color("240") // border gray
	green    = lipgloss.Color("46")  // attached indicator
	red      = lipgloss.Color("196") // errors
	inputBg  = lipgloss.Color("237") // input background
)

// --- Styles ---

var (
	frameStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(border).
			Padding(0, 1)

	titleStyle = lipgloss.NewStyle().
			Foreground(accent).
			Bold(true)

	sessionStyle = lipgloss.NewStyle().
			Foreground(text)

	selectedStyle = lipgloss.NewStyle().
			Foreground(accent).
			Background(inputBg).
			PaddingLeft(1).
			PaddingRight(1)

	attachedDot = lipgloss.NewStyle().
			Foreground(green).
			Bold(true)

	detachedDot = lipgloss.NewStyle().
			Foreground(dim)

	helpStyle = lipgloss.NewStyle().
			Foreground(dim)

	tipStyle = lipgloss.NewStyle().
			Foreground(dim).
			Italic(true)

	inputStyle = lipgloss.NewStyle().
			Foreground(text).
			Background(inputBg).
			PaddingLeft(1).
			PaddingRight(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(red).
			Bold(true)

	dimText = lipgloss.NewStyle().
		Foreground(dim)
)

// View renders the current model state.
func (m *Model) View() string {
	if m.quitting {
		return ""
	}

	var content string
	switch m.mode {
	case ModeList:
		content = m.viewList()
	case ModeCreate:
		content = m.viewInput("Create new session:", "[enter] create  [esc] cancel")
	case ModeRename:
		label := "Rename session:"
		if len(m.sessions) > 0 && m.cursor < len(m.sessions) {
			label = fmt.Sprintf("Rename '%s':", m.sessions[m.cursor].Name)
		}
		content = m.viewInput(label, "[enter] rename  [esc] cancel")
	case ModeConfirmKill:
		content = m.viewConfirm()
	}

	w := max(50, m.width-4)
	return frameStyle.Width(w).Render(content)
}

// viewList renders the session list with help bar.
func (m *Model) viewList() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("tmux-pilot"))
	b.WriteString("\n\n")

	if len(m.sessions) == 0 {
		b.WriteString(dimText.Render("  No sessions found"))
	} else {
		for i, s := range m.sessions {
			b.WriteString(formatSession(s, i == m.cursor))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  [enter] switch  [n] new  [r] rename"))
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("  [x] kill  [q] quit"))
	b.WriteString("\n\n")
	b.WriteString(tipStyle.Render("  tip: Ctrl-b d to detach from tmux"))

	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render("  Error: " + m.err.Error()))
	}

	return b.String()
}

// viewInput renders create/rename prompts.
func (m *Model) viewInput(label, help string) string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("tmux-pilot"))
	b.WriteString("\n\n")
	b.WriteString("  " + label + "\n\n")
	b.WriteString("  " + inputStyle.Render(m.input+"█"))
	b.WriteString("\n\n")
	b.WriteString(helpStyle.Render("  " + help))

	if m.err != nil {
		b.WriteString("\n\n")
		b.WriteString(errorStyle.Render("  Error: " + m.err.Error()))
	}

	return b.String()
}

// viewConfirm renders the kill confirmation dialog.
func (m *Model) viewConfirm() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("tmux-pilot"))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "  Kill session '%s'?\n\n", m.killName)
	b.WriteString(helpStyle.Render("  [y/enter] yes  [n/esc] no"))

	return b.String()
}

// formatSession renders a single session line.
func formatSession(s tmux.Session, selected bool) string {
	// Dot indicator
	var dot string
	if s.Attached {
		dot = attachedDot.Render("●")
	} else {
		dot = detachedDot.Render("○")
	}

	// Window count
	win := fmt.Sprintf("%d window", s.WindowCount)
	if s.WindowCount != 1 {
		win += "s"
	}

	// Status
	status := "detached"
	if s.Attached {
		status = "attached"
	}

	info := fmt.Sprintf("%-15s %s   %s", s.Name, win, status)

	if selected {
		return "  " + dot + " " + selectedStyle.Render(info)
	}
	return "  " + dot + " " + sessionStyle.Render(info)
}
