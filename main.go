package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/damonto/estkme-cloud/internal/cloud"
	"github.com/damonto/estkme-cloud/internal/config"
	"github.com/damonto/estkme-cloud/internal/lpac"
)

var Version string

func init() {
	if err := os.MkdirAll("/tmp/estkme-cloud", 0755); err != nil {
		panic(err)
	}
	flag.StringVar(&config.C.ListenAddress, "listen-address", ":1888", "eSTK.me cloud enhance server listen address")
	flag.StringVar(&config.C.Dir, "dir", "/tmp/estkme-cloud", "the directory to store lpac")
	flag.StringVar(&config.C.Version, "version", "v2.1.0", "the version of lpac to download")
	flag.BoolVar(&config.C.DontDownload, "dont-download", false, "don't download lpac")
	flag.StringVar(&config.C.Advertising, "advertising", "", "advertising message to show on the server (max: 100 characters)")
	flag.BoolVar(&config.C.Verbose, "verbose", false, "verbose mode")
	flag.Parse()
}

func initApp() {
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
		if err := lpac.Download(config.C.Dir, config.C.Version); err != nil {
			slog.Error("failed to download lpac", "error", err)
			os.Exit(1)
		}
	}
}

func main() {
	slog.Info("eSTK.me cloud enhance server", "version", Version)
	initApp()

	manager := cloud.NewManager()
	server := cloud.NewServer(manager)

	go func() {
		if err := server.Listen(config.C.ListenAddress); err != nil {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGHUP, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	<-quit
	slog.Info("shutting down server")
	if err := server.Shutdown(); err != nil {
		slog.Error("failed to shutdown server", "error", err)
		os.Exit(1)
	}
}
