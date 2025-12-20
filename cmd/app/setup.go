package main

import (
	"embed"
	"fmt"
	"log/slog"
	"time"

	"github.com/bit8bytes/goalkeepr/internal/branding"
	"github.com/bit8bytes/goalkeepr/internal/database"
	"github.com/bit8bytes/goalkeepr/internal/flags"
	"github.com/bit8bytes/goalkeepr/internal/goals"
	"github.com/bit8bytes/goalkeepr/internal/logger"
	"github.com/bit8bytes/goalkeepr/internal/share"
	"github.com/bit8bytes/goalkeepr/internal/users"

	"github.com/alexedwards/scs/sqlite3store"
	"github.com/alexedwards/scs/v2"
)

//go:embed "db/migrations"
var migrations embed.FS

const (
	// GoalkeeprCookie is the session cookie name used across the application.
	GoalkeeprCookie = "goalkeepr"

	HTMLDateFormat = "2006-01-02"
)

type services struct {
	users    *users.Service
	goals    *goals.Service
	branding *branding.Service
	share    *share.Service
}

// newApp creates and initializes a new app instance with the given configuration.
// It sets up logging, database, session management, and all application services.
func newApp(cfg *flags.Options) (*app, error) {
	// Initialize logger with variable log level
	var level = new(slog.LevelVar)
	logger := logger.Setup(level)

	if cfg.Env.IsDev() {
		level.Set(slog.LevelDebug)
	} else {
		level.Set(slog.LevelInfo) // Production: Info level for operational visibility
	}

	templateCache, err := newTemplateCache()
	if err != nil {
		return nil, fmt.Errorf("template cache failure: %w", err)
	}

	db, err := database.Open(cfg.Database.Driver, cfg.Database.Dsn)
	if err != nil {
		return nil, fmt.Errorf("open database failure: %w", err)
	}

	dbVersion, err := database.Migrate(db, migrations)
	if err != nil {
		return nil, fmt.Errorf("migrate database failure: %w", err)
	}

	if err := database.Verify(db); err != nil {
		return nil, fmt.Errorf("database verification failure: %w", err)
	}

	if err := database.Optimize(db); err != nil {
		return nil, fmt.Errorf("database optimization failure: %w", err)
	}

	logger.Info("database",
		"version", dbVersion,
		"driver", cfg.Database.Driver,
		"verify", "ok",
		"optimize", "done",
	)

	// Configure session manager with SQLite store
	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Name = GoalkeeprCookie
	sessionManager.Store = sqlite3store.New(db)

	// q := &queries{}

	services := &services{
		users:    users.NewService(db),
		goals:    goals.NewService(db),
		branding: branding.NewService(db),
		share:    share.NewService(db),
	}

	app := &app{
		config:         cfg,
		logger:         logger,
		templateCache:  templateCache,
		sessionManager: sessionManager,
		services:       services,
		limiters:       newLimiters(),
	}

	return app, nil
}
