package app

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const responseTime = 8500 * time.Millisecond // The amount of milliseconds for the result to be displayed

func (m Model) Init() tea.Cmd {
	return nil
}
