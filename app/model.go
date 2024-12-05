package app

import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
)

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
