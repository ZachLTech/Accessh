package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/log"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

type Location struct {
	Servicename string `json:"service"`
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
	textInput   textinput.Model
	config      Config
	err         error
	destination string
	showHelp    bool
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
	return m, []tea.ProgramOption{tea.WithInput(os.Stdin)}
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

	p := tea.NewProgram(initialModel)
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
	ti := textinput.New()
	ti.Placeholder = "Enter where you're trying to go (e.g., exit.zachl.tech, zachl.tech)"
	ti.Focus()
	return Model{
		textInput: ti,
		config:    config,
		err:       err,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tickMsg:
		m.textInput.Placeholder = "Enter where you're trying to go (e.g., exit.zachl.tech, zachl.tech)"

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			m.destination = m.textInput.Value()
			if location, exists := m.config.Locations[m.destination]; exists {
				m.textInput.Placeholder = fmt.Sprintf("Run this command to access %v: ssh -p %v %v\n", location.Servicename, location.Port, location.Hostname)
				m.textInput.SetValue("")
				return m, tickResponse()
			} else {
				m.textInput.Placeholder = fmt.Sprintf("no destination found at location: %s", m.destination)
				m.textInput.SetValue("")
				return m, tickResponse()
			}
		}
	}
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func getLocations(locations map[string]Location) string {

	for _, location := range locations {

	}

}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("Error: %v\n\n%s\n\n(esc to quit)\n\n", m.err, m.textInput.View())
	}

	if m.showHelp {
		return getLocations(m.config.Locations)
	}

	return fmt.Sprintf("\n\n%v\n\n%v:\n\n%s\n\n(esc to quit)", m.config.Settings.Title, m.config.Settings.Description, m.textInput.View())
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
