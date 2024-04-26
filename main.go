package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/damonto/estkme-cloud/internal/cloud"
	"github.com/damonto/estkme-cloud/internal/config"
	"github.com/damonto/estkme-cloud/internal/lpac"
)

var Version string

func init() {
	cwd, _ := os.Getwd()
	flag.StringVar(&config.C.ListenAddress, "listen-address", ":1888", "eSTK.me cloud enhance server listen address")
	flag.StringVar(&config.C.DataDir, "data-dir", filepath.Join(cwd, "data"), "data directory")
	flag.StringVar(&config.C.LpacVersion, "lpac-version", "v2.0.0", "lpac version")
	flag.BoolVar(&config.C.DontDownload, "dont-download", false, "don't download lpac")
	flag.BoolVar(&config.C.Verbose, "verbose", false, "verbose mode")
	flag.Parse()
}

func main() {
	slog.Info("eSTK.me cloud enhance server", "version", Version)
	config.C.LoadEnv()

	if config.C.Verbose {
		slog.SetLogLoggerLevel(slog.LevelDebug)
		slog.Warn("verbose mode is enabled, this will print out sensitive information")
	}

	if err := config.C.IsValid(); err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	if !config.C.DontDownload {
		if err := lpac.Download(config.C.DataDir, config.C.LpacVersion); err != nil {
			slog.Error("failed to download lpac", "error", err)
			os.Exit(1)
		}
	}

	manager := cloud.NewManager()
	server := cloud.NewServer(manager)

	go func() {
		if err := server.Listen(config.C.ListenAddress); err != nil {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	slog.Info("shutting down server")
	if err := server.Shutdown(); err != nil {
		slog.Error("failed to shutdown server", "error", err)
		os.Exit(1)
	}
}
