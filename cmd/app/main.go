package main

import (
	"html/template"
	"log"
	"log/slog"
	"sync"

	"github.com/alexedwards/scs/v2"
	"github.com/bit8bytes/goalkeepr/internal/flags"
	_ "modernc.org/sqlite"
)

type app struct {
	config         *flags.Options
	logger         *slog.Logger
	templateCache  map[string]*template.Template
	sessionManager *scs.SessionManager
	services       *services
	limiters       *limiters

	wg sync.WaitGroup
}

func main() {
	cfg, err := flags.Parse()
	if err != nil {
		log.Fatalf("parse flags failure: %v", err)
	}

	app, err := newApp(cfg)
	if err != nil {
		log.Fatalf("setup failure: %v", err)
	}

	if err := app.serve(); err != nil {
		log.Fatalf("server failure: %v", err)
	}
}
