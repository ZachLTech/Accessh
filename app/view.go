package app

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/wordwrap"
)

func (m Model) View() string {
	maxWidth := m.width - 4

	if m.err != nil {
		errorMsg := fmt.Sprintf("Error: %v", m.err)
		return lipgloss.NewStyle().PaddingLeft(2).PaddingRight(2).Render(wordwrap.String(errorMsg, maxWidth) + "\n\n" + m.textInput.View() + "\n\n(Ctrl+C to quit)\n\n")
	}
	if m.showHelp {
		menu := fmt.Sprintf("Public Services (Esc to go back):\n\n%v\n\n↑/↓: Navigate • Esc: Go back", m.viewport.View())
		return lipgloss.NewStyle().PaddingLeft(2).PaddingRight(2).PaddingTop(1).PaddingBottom(1).Render(wordwrap.String(menu, maxWidth))
	}

	mainContent := fmt.Sprintf("\n%v\n\n%v:\n\n%s\n\n(Ctrl+C to quit)", m.config.Settings.Title, m.config.Settings.Description, m.textInput.View())

	return lipgloss.NewStyle().PaddingLeft(2).PaddingRight(2).Render(wordwrap.String(mainContent, maxWidth))
}
