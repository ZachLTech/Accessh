package main

import (
	"Accessh/app"

	"github.com/charmbracelet/log"
)

func main() {
	config, err := app.LoadConfig("config.json")
	if err != nil {
		log.Error("Could not load config", "error", err)
	}

	if config.Settings.SSHSettings.Enabled {
		app.RunSSHServer(config)
	} else {
		app.RunLocalTUI()
	}
}
