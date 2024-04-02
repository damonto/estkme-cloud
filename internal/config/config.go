package config

import (
	"errors"
	"net"
)

type Config struct {
	ListenAddress string
	LpacVersion   string
	DataDir       string
	DontDownload  bool
}

var C = &Config{}

var (
	ErrLpacVersionRequired = errors.New("lpac version is required")
)

func (c *Config) IsValid() error {
	if _, err := net.ResolveTCPAddr("tcp", c.ListenAddress); err != nil {
		return err
	}
	if c.LpacVersion == "" {
		return ErrLpacVersionRequired
	}
	return nil
}
