package tui

import (
	"fmt"
	"strings"

	"github.com/blockful/tmux-pilot/internal/tmux"
	"github.com/charmbracelet/lipgloss"
)

var (
	accent  = lipgloss.Color("205")
	text    = lipgloss.Color("252")
	dim     = lipgloss.Color("243")
	border  = lipgloss.Color("240")
	green   = lipgloss.Color("46")
	yellow  = lipgloss.Color("214")
	inputBg = lipgloss.Color("237")

	frame     = lipgloss.NewStyle().BorderStyle(lipgloss.RoundedBorder()).BorderForeground(border).Padding(0, 1)
	title     = lipgloss.NewStyle().Foreground(accent).Bold(true)
	selected  = lipgloss.NewStyle().Foreground(accent).Background(inputBg).PaddingLeft(1).PaddingRight(1)
	normal    = lipgloss.NewStyle().Foreground(text)
	dotOn     = lipgloss.NewStyle().Foreground(green).Bold(true)
	dotOff    = lipgloss.NewStyle().Foreground(dim)
	dimStyle  = lipgloss.NewStyle().Foreground(dim)
	warnStyle = lipgloss.NewStyle().Foreground(yellow).Bold(true)
	inputShow = lipgloss.NewStyle().Foreground(text).Background(inputBg).PaddingLeft(1).PaddingRight(1)
	tipStyle  = lipgloss.NewStyle().Foreground(dim).Italic(true)
)

// View renders the UI.
func (m *Model) View() string {
	var content string
	switch m.mode {
	case ModeList:
		content = m.viewList()
	case ModeCreate:
		content = m.viewInput("New session name:")
	case ModeRename:
		content = m.viewInput(fmt.Sprintf("Rename '%s':", m.sessions[m.cursor].Name))
	case ModeConfirmKill:
		content = m.viewConfirm()
	}
	return frame.Width(max(50, m.width-4)).Render(content)
}

func (m *Model) viewList() string {
	var b strings.Builder
	b.WriteString(title.Render("tmux-pilot"))
	b.WriteString("\n\n")

	if len(m.sessions) == 0 {
		b.WriteString(dimStyle.Render("  No sessions"))
	} else {
		for i, s := range m.sessions {
			b.WriteString(fmtSession(s, i == m.cursor))
			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  [enter] switch  [n] new  [r] rename"))
	b.WriteString("\n")
	b.WriteString(dimStyle.Render("  [x] kill  [d] detach  [q] quit"))
	b.WriteString("\n\n")
	b.WriteString(tipStyle.Render("  tip: Ctrl-b d to detach"))
	return b.String()
}

func (m *Model) viewInput(label string) string {
	var b strings.Builder
	b.WriteString(title.Render("tmux-pilot"))
	b.WriteString("\n\n")
	b.WriteString("  " + label + "\n\n")
	b.WriteString("  " + inputShow.Render(m.input+"█"))
	if m.warning != "" {
		b.WriteString("\n\n")
		b.WriteString(warnStyle.Render("  ⚠ " + m.warning))
	}
	b.WriteString("\n\n")
	b.WriteString(dimStyle.Render("  [enter] confirm  [esc] cancel"))
	return b.String()
}

func (m *Model) viewConfirm() string {
	var b strings.Builder
	b.WriteString(title.Render("tmux-pilot"))
	b.WriteString("\n\n")
	fmt.Fprintf(&b, "  Kill session '%s'?\n\n", m.sessions[m.cursor].Name)
	b.WriteString(dimStyle.Render("  [y/enter] yes  [n/esc] no"))
	return b.String()
}

func fmtSession(s tmux.Session, sel bool) string {
	dot := dotOff.Render("○")
	if s.Attached {
		dot = dotOn.Render("●")
	}
	win := fmt.Sprintf("%d window", s.WindowCount)
	if s.WindowCount != 1 {
		win += "s"
	}
	status := "detached"
	if s.Attached {
		status = "attached"
	}
	info := fmt.Sprintf("%-15s %s   %s", s.Name, win, status)
	if sel {
		return "  " + dot + " " + selected.Render(info)
	}
	return "  " + dot + " " + normal.Render(info)
}
