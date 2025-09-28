package main

import (
	"flag"
	"html/template"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/bit8bytes/goalkeepr/internal/branding"
	"github.com/bit8bytes/goalkeepr/internal/data"
	"github.com/bit8bytes/goalkeepr/internal/goals"
	"github.com/bit8bytes/goalkeepr/internal/share"
	"github.com/bit8bytes/goalkeepr/internal/users"
)

const (
	UserIDSessionKey    = "userID"
	HTMLDateFormat      = "2006-01-02"
	GoalkeeprCookieName = "goalkeepr"
)

type config struct {
	env  string
	port int
	db   data.Database
}

type app struct {
	config         config
	logger         *slog.Logger
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
	modules        *modules

	wg sync.WaitGroup
}

type modules struct {
	users    *users.Service
	goals    *goals.Service
	branding *branding.Service
	share    *share.Service
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

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Name = GoalkeeprCookieName

	modules := &modules{
		users:    users.New(dbP.DB),
		goals:    goals.New(dbP.DB),
		branding: branding.New(dbP.DB),
		share:    share.New(dbP.DB),
	}

	app := app{
		config:         cfg,
		logger:         logger,
		templateCache:  templateCache,
		sessionManager: sessionManager,
		modules:        modules,
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
