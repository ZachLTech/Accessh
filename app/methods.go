package app

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
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
)

/************************************ Runner functions **************************************/
/************************ runs either the local game or SSH server **************************/

func RunLocalTUI() {
	initialModel := initialModel()

	p := tea.NewProgram(initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v", err)
		os.Exit(1)
	}
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

/*********************************** SSH Server Handler *************************************/

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	m := initialModel()
	return m, []tea.ProgramOption{tea.WithInput(os.Stdin), tea.WithAltScreen()}
}

/**************************************** Tickers *******************************************/

func tickResponse() tea.Cmd {
	return tea.Tick(responseTime, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

/************************************* Helpers & Utils **************************************/

func LoadConfig(filename string) (Config, error) {
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
	config, err := LoadConfig("config.json")
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

func getLocations(locations map[string]Location) string {
	var builder strings.Builder
	for _, location := range locations {
		locationCard := fmt.Sprintf("%v\nLocation: %v\nDescription: %v\nSource: %v\n\n", location.Servicename, location.Hostname, location.Description, location.Repository)
		builder.WriteString(locationCard)
	}
	return builder.String()
}
