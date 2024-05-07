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
	Advertising   string
	Verbose       bool
}

var C = &Config{}

var (
	ErrLpacVersionRequired = errors.New("lpac version is required")
	ErrAdvertisingTooLong  = errors.New("advertising message is too long (max: 100 characters)")
	ErrInvalidAdvertising  = errors.New("advertising message contains non-printable ASCII characters")
)

func (c *Config) IsValid() error {
	if _, err := net.ResolveTCPAddr("tcp", c.ListenAddress); err != nil {
		return err
	}
	if c.LpacVersion == "" {
		return ErrLpacVersionRequired
	}
	if len(c.Advertising) > 100 {
		return ErrAdvertisingTooLong
	}
	// Advertising message is only allowed contain printable ASCII characters
	for _, r := range c.Advertising {
		if r < 32 || r > 126 {
			return ErrInvalidAdvertising
		}
	}
	return nil
}

func (c *Config) GetAdvertising() string {
	if c.Advertising != "" {
		return "!! Advertising !! \n" + c.Advertising
	}
	return c.Advertising
}

func (c *Config) LoadEnv() {
	if os.Getenv("ESTKME_CLOUD_LISTEN_ADDRESS") != "" {
		c.ListenAddress = os.Getenv("ESTKME_CLOUD_LISTEN_ADDRESS")
	}
	if os.Getenv("ESTKME_CLOUD_LPAC_VERSION") != "" {
		c.LpacVersion = os.Getenv("ESTKME_CLOUD_LPAC_VERSION")
	}
	if os.Getenv("ESTKME_CLOUD_DATA_DIR") != "" {
		c.DataDir = os.Getenv("ESTKME_CLOUD_DATA_DIR")
	}
	if os.Getenv("ESTKME_CLOUD_DONT_DOWNLOAD") != "" {
		c.DontDownload = true
	}
	if os.Getenv("ESTKME_CLOUD_ADVERTISING") != "" {
		c.Advertising = os.Getenv("ESTKME_CLOUD_ADVERTISING")
	}
	if os.Getenv("ESTKME_CLOUD_VERBOSE") != "" {
		c.Verbose = true
	}
}
