package config

import (
	"net"
)

type Config struct {
	ListenAddress string
	LpacVersion   string
	DataDir       string
	BotToken      string
}

var C = &Config{}

func (c *Config) IsValid() error {
	if _, err := net.ResolveTCPAddr("tcp", c.ListenAddress); err != nil {
		return err
	}

	return nil
}
