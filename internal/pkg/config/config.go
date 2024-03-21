package config

import (
	"errors"
	"net"
)

type Config struct {
	AppListenAddress string
	ListenAddress    string
	LpacVersion      string
	DataDir          string
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

	if _, err := net.ResolveTCPAddr("tcp", c.AppListenAddress); err != nil {
		return err
	}

	return nil
}
