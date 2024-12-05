package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/charmbracelet/log"
	"golang.org/x/term"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"github.com/muesli/reflow/wordwrap"
)

type Location struct {
	Servicename string `json:"service"`
	Description string `json:"description"`
	Repository  string `json:"repo"`
	Hostname    string `json:"hostname"`
	Port        int    `json:"port"`
}

type SSHSettings struct {
	Enabled  bool   `json:"enabled"`
	Hostname string `json:"hostname"`
	Port     string `json:"port"`
}

type ProgramSettings struct {
	Title       string      `json:"title"`
	Description string      `json:"description"`
	SSHSettings SSHSettings `json:"SSH"`
}

type Config struct {
	Settings  ProgramSettings     `json:"settings"`
	Locations map[string]Location `json:"locations"`
}

type Model struct {
	textInput     textinput.Model
	locationCards string
	config        Config
	err           error
	destination   string
	showHelp      bool
	width         int
	viewport      viewport.Model
}

const responseTime = 5000 * time.Millisecond // The amount of milliseconds for the result to be displayed

type tickMsg time.Time

func tickResponse() tea.Cmd {
	return tea.Tick(responseTime, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	m := initialModel()
	return m, []tea.ProgramOption{tea.WithInput(os.Stdin), tea.WithAltScreen()}
}

func RunSSHServer(config Config) {
	var (
		host = config.Settings.SSHSettings.Hostname
		port = config.Settings.SSHSettings.Port
	)

	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),

		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}

func RunLocalTUI() {
	initialModel := initialModel()

	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}

func loadConfig(filename string) (Config, error) {
	var config Config
	file, err := os.Open(filename)
	if err != nil {
		return config, fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&config); err != nil {
		return config, fmt.Errorf("error decoding config: %v", err)
	}
	return config, nil
}

func initialModel() Model {
	config, err := loadConfig("config.json")
	width, height, _ := term.GetSize(0)
	textInput := textinput.New()
	textInput.Placeholder = "Enter where you're trying to go (e.g., exit.zachl.tech, zachl.tech)"
	textInput.Focus()

	viewport := viewport.New(width, height-6) // Leave space for header/footer
	viewport.SetContent("Loading...")

	return Model{
		textInput:     textInput,
		locationCards: getLocations(config.Locations),
		config:        config,
		err:           err,
		width:         width,
		viewport:      viewport,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

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

func getLocations(locations map[string]Location) string {
	var builder strings.Builder
	for _, location := range locations {
		locationCard := fmt.Sprintf("%v\nLocation: %v\nDescription: %v\nSource: %v\n\n", location.Servicename, location.Hostname, location.Description, location.Repository)
		builder.WriteString(locationCard)
	}
	return builder.String()
}

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

func main() {
	config, err := loadConfig("config.json")
	if err != nil {
		log.Error("Could not load config", "error", err)
	}

	if config.Settings.SSHSettings.Enabled {
		RunSSHServer(config)
	} else {
		RunLocalTUI()
	}
}
