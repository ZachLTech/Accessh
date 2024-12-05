package app

import "time"

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

type tickMsg time.Time
