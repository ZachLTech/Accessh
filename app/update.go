package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tickMsg:
		m.textInput.Placeholder = "Enter where you're trying to go (e.g., exit.zachl.tech, zachl.tech)"
	case tea.WindowSizeMsg:
		cmd = tea.ClearScreen
		m.width = msg.Width
		m.viewport.Width = msg.Width
		m.viewport.Height = msg.Height - 9
		return m, cmd
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEsc:
			if m.showHelp {
				m.textInput.SetValue("")
				m.showHelp = false
				return m, cmd
			} else {
				return m, cmd
			}
		case tea.KeyEnter:
			m.destination = m.textInput.Value()
			if m.destination == "help" {
				m.showHelp = true
				m.viewport.SetContent(m.locationCards)
				m.viewport.GotoTop()
				return m, cmd
			}
			if location, exists := m.config.Locations[m.destination]; exists {
				m.textInput.Placeholder = fmt.Sprintf("Run this command to access %v: ssh -p %v %v\n", location.Servicename, location.Port, location.Hostname)
				m.textInput.SetValue("")
				return m, tickResponse()
			} else {
				m.textInput.Placeholder = fmt.Sprintf("no destination found at location: %s", m.destination)
				m.textInput.SetValue("")
				return m, tickResponse()
			}
		case tea.KeyUp, tea.KeyDown:
			if m.showHelp {
				m.viewport, cmd = m.viewport.Update(msg)
				return m, cmd
			}
		}
	}

	if m.showHelp {
		viewport, cmd := m.viewport.Update(msg)
		m.viewport = viewport
		cmds = append(cmds, cmd)
	}

	m.textInput, cmd = m.textInput.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}
