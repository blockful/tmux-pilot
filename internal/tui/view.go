package tui

// View renders the model. Placeholder — full implementation in next PR.
func (m *Model) View() string {
	if m.quitting {
		return ""
	}
	return "tmux-pilot"
}
