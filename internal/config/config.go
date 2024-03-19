package config

import (
	"errors"
)

type Config struct {
	ListenAddress string
	LpacVersion   string
	DataDir       string
	BotToken      string
}

var C = &Config{}

func (c *Config) IsValid() error {
	if c.BotToken == "" {
		return errors.New("bot token is required")
	}
	return nil
}
