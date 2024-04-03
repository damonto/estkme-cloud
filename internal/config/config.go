package config

import (
	"errors"
	"net"
	"os"
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

func (c *Config) LoadEnv() {
	if os.Getenv("RLPA_LISTEN_ADDRESS") != "" {
		c.ListenAddress = os.Getenv("RLPA_LISTEN_ADDRESS")
	}
	if os.Getenv("RLPA_LPAC_VERSION") != "" {
		c.LpacVersion = os.Getenv("RLPA_LPAC_VERSION")
	}
	if os.Getenv("RLPA_DATA_DIR") != "" {
		c.DataDir = os.Getenv("RLPA_DATA_DIR")
	}
	if os.Getenv("RLPA_DONT_DOWNLOAD") != "" {
		c.DontDownload = true
	}
}
