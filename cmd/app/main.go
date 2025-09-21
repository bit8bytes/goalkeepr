package main

import (
	"flag"
	"log/slog"
	"os"
	"sync"
)

type config struct {
	env  string
	port int
}

type app struct {
	config config
	logger *slog.Logger

	wg sync.WaitGroup
}

func main() {
	var logLevel = new(slog.LevelVar)
	loggerOpts := &slog.HandlerOptions{Level: logLevel}
	logger := slog.New(slog.NewTextHandler(os.Stdout, loggerOpts))

	var cfg config
	if err := setup(&cfg); err != nil {
		logger.Error("Error occured on setup", "err", err)
		os.Exit(1)
	}

	app := app{
		config: cfg,
		logger: logger,
	}

	if err := app.serve(); err != nil {
		logger.Error("Error shutting down the server", "err", err)
		os.Exit(1)
	}
}

func setup(cfg *config) error {
	flag.StringVar(&cfg.env, "env", "prod", "Environment (dev|stage|prod)")
	flag.IntVar(&cfg.port, "port", 8080, "Port")

	flag.Parse()

	return nil
}
