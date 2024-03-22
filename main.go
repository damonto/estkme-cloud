package main

import (
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"

	"github.com/damonto/estkme-rlpa-server/internal/config"
	"github.com/damonto/estkme-rlpa-server/internal/lpac"
	"github.com/damonto/estkme-rlpa-server/internal/rlpa"
)

func init() {
	cwd, _ := os.Getwd()
	flag.StringVar(&config.C.ListenAddress, "listen-address", ":1888", "rLPA server listen address")
	flag.StringVar(&config.C.LpacVersion, "lpac-version", "v2.0.0-beta.1", "lpac version")
	flag.StringVar(&config.C.DataDir, "data-dir", filepath.Join(cwd, "data"), "data directory")
	flag.Parse()
}

func main() {
	if err := config.C.IsValid(); err != nil {
		slog.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	if err := lpac.Download(config.C.DataDir, config.C.LpacVersion); err != nil {
		slog.Error("failed to download lpac", "error", err)
		os.Exit(1)
	}

	manager := rlpa.NewManager()
	server := rlpa.NewServer(manager)

	go func() {
		if err := server.Listen(config.C.ListenAddress); err != nil {
			panic(err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	slog.Info("shutting down server")
	server.Shutdown()
}
