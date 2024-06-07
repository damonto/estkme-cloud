package config

import (
	"errors"
	"net"
	"os"
	"strings"
)

type Config struct {
	ListenAddress string
	Version       string
	Dir           string
	DontDownload  bool
	Advertising   string
	Verbose       bool
}

var C = &Config{}

var (
	ErrVersionRequired    = errors.New("lpac version is required")
	ErrAdvertisingTooLong = errors.New("advertising message is too long (max: 100 characters)")
	ErrInvalidAdvertising = errors.New("advertising message contains non-printable ASCII characters")
)

func (c *Config) IsValid() error {
	if _, err := net.ResolveTCPAddr("tcp", c.ListenAddress); err != nil {
		return err
	}
	if c.Version == "" {
		return ErrVersionRequired
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

func (c *Config) GetAdvertising() []byte {
	if c.Advertising != "" {
		return []byte("!! Advertising !! \n" + strings.Replace(c.Advertising, "_br_", "\n", -1))
	}
	return []byte{}
}

func (c *Config) LoadEnv() {
	if os.Getenv("ESTKME_CLOUD_LISTEN_ADDRESS") != "" {
		c.ListenAddress = os.Getenv("ESTKME_CLOUD_LISTEN_ADDRESS")
	}
	if os.Getenv("ESTKME_CLOUD_LPAC_VERSION") != "" {
		c.Version = os.Getenv("ESTKME_CLOUD_LPAC_VERSION")
	}
	if os.Getenv("ESTKME_CLOUD_DATA_DIR") != "" {
		c.Dir = os.Getenv("ESTKME_CLOUD_DATA_DIR")
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
