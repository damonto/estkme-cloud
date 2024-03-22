package config

import (
	"errors"
	"net"
)

type Config struct {
	ListenAddress    string
	LpacVersion      string
	DataDir          string
	BotToken		 string
}

var C = &Config{}

var (
	ErrLpacVersionEmpty = errors.New("lpac version is empty")
)

func (c *Config) IsValid() error {
	if _, err := net.ResolveTCPAddr("tcp", c.ListenAddress); err != nil {
		return err
	}
	if c.LpacVersion == "" {
		return ErrLpacVersionEmpty
	}
	return nil
}
