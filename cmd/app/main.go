package main

import (
	"flag"
	"html/template"
	"log/slog"
	"os"
	"sync"

	"github.com/bit8bytes/goalkeepr/internal/data"
)

type config struct {
	env  string
	port int
	db   data.Database
}

type app struct {
	config        config
	logger        *slog.Logger
	templateCache map[string]*template.Template

	wg sync.WaitGroup
}

func main() {
	var logLevel = new(slog.LevelVar)
	loggerOpts := &slog.HandlerOptions{Level: logLevel}
	logger := slog.New(slog.NewTextHandler(os.Stdout, loggerOpts))

	var cfg config
	if err := setup(&cfg); err != nil {
		logger.Error("Error occured on setup", slog.String("msg", err.Error()))
		os.Exit(1)
	}

	if cfg.env == "dev" {
		logLevel.Set(slog.LevelDebug)
	}

	dbP, err := data.New(cfg.db.Driver, cfg.db.Path)
	if err != nil {
		logger.Error("Error creating database provider", slog.String("msg", err.Error()))
		os.Exit(1)
	}

	dbVersion, err := dbP.AutoMigrate()
	if err != nil {
		logger.Error("Error migrating",
			slog.String("msg", err.Error()),
			slog.String("driver", dbP.Driver),
			slog.String("path", dbP.Path),
		)
		os.Exit(1)
	}

	logger.Info("Database version", slog.Int64("version", *dbVersion))

	templateCache, err := newTemplateCache()
	if err != nil {
		logger.Error("Error creating template cache", slog.String("msg", err.Error()))
		os.Exit(1)
	}

	app := app{
		config:        cfg,
		templateCache: templateCache,
		logger:        logger,
	}

	if err := app.serve(); err != nil {
		logger.Error("Error shutting down the server", slog.String("msg", err.Error()))
		os.Exit(1)
	}
}

func setup(cfg *config) error {
	flag.StringVar(&cfg.env, "env", "prod", "Environment (dev|stage|prod)")
	flag.IntVar(&cfg.port, "port", 8080, "Port")

	// Database configuration
	flag.StringVar(&cfg.db.Driver, "database-driver", "sqlite", "database driver")
	flag.StringVar(&cfg.db.Path, "database-path", "", "database path")

	flag.Parse()

	return nil
}
