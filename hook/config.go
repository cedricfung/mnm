package main

import (
	"os"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Mixin struct {
		ClientID   string `toml:"client-id"`
		SessionID  string `toml:"session-id"`
		PrivateKey string `toml:"private-key"`
	} `toml:"mixin"`
	App struct {
		Port int `toml:"port"`
	} `toml:"app"`
}

func LoadConfig(file string) (*Config, error) {
	f, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	var config Config
	err = toml.Unmarshal(f, &config)
	return &config, err
}
