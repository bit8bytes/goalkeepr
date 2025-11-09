package main

import (
	"flag"
	"fmt"
	"html"
	"html/template"
	"log/slog"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/alexedwards/scs/v2"
	_ "modernc.org/sqlite"

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
	limiters       *limiters
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
	logger := setupLogger(logLevel)

	var cfg config
	if err := setupConfig(&cfg); err != nil {
		logger.Error("error occured on setup", slog.String("msg", err.Error()))
		os.Exit(1)
	}

	if cfg.env == "dev" {
		logLevel.Set(slog.LevelDebug)
	}

	dbP, err := data.New(cfg.db.Driver, cfg.db.Path)
	if err != nil {
		logger.Error("error creating database provider", slog.String("msg", err.Error()))
		os.Exit(1)
	}

	dbVersion, err := dbP.AutoMigrate()
	if err != nil {
		logger.Error("error migrating",
			slog.String("msg", err.Error()),
			slog.String("driver", dbP.Driver),
			slog.String("path", dbP.Path),
		)
		os.Exit(1)
	}

	logger.Info("database version", slog.Int64("version", *dbVersion))

	htmxFuncs := template.FuncMap{
		"preload": func(event string) template.HTMLAttr {
			escaped := html.EscapeString(event)
			return template.HTMLAttr(fmt.Sprintf(`preload="%s"`, escaped))
		},
	}

	templateCache, err := newTemplateCache(WithFunctions(htmxFuncs))
	if err != nil {
		logger.Error("error creating template cache", slog.String("msg", err.Error()))
		os.Exit(1)
	}

	sessionManager := scs.New()
	sessionManager.Lifetime = 24 * time.Hour
	sessionManager.Cookie.Name = GoalkeeprCookieName

	limiters := newLimiters()

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
		limiters:       limiters,
	}

	if err := app.serve(); err != nil {
		logger.Error("error shutting down the server", slog.String("msg", err.Error()))
		os.Exit(1)
	}
}

func setupConfig(cfg *config) error {
	flag.StringVar(&cfg.env, "env", "prod", "Environment (dev|stage|prod)")
	flag.IntVar(&cfg.port, "port", 8080, "Port")

	// Database configuration
	flag.StringVar(&cfg.db.Driver, "database-driver", "sqlite", "database driver")
	flag.StringVar(&cfg.db.Path, "database-path", "", "database path")

	flag.Parse()

	envs := []string{"dev", "stage", "prod"}
	if ok := slices.Contains(envs, cfg.env); !ok {
		return fmt.Errorf("env must be on of: %v", envs)
	}

	if cfg.port < 0 || cfg.port > 65535 {
		return fmt.Errorf("port is not in valid range of 0-65535")
	}

	return nil
}

func setupLogger(level *slog.LevelVar) *slog.Logger {
	loggerOpts := &slog.HandlerOptions{AddSource: false, Level: level}
	handler := slog.NewTextHandler(os.Stdout, loggerOpts)
	traceHandler := newTraceHandler(handler)

	return slog.New(traceHandler)
}
