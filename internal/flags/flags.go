// Package flags provides command-line flag parsing.
package flags

import (
	"flag"
	"fmt"
	"slices"
)

// Options contains command-line configuration.
type Options struct {
	Env      Env
	Port     int
	Database struct {
		Driver string
		Dsn    string
	}
}

// Parse parses command-line flags and returns Options.
func Parse() (*Options, error) {
	cfg := &Options{}
	cfg.Env.slug = "prod"

	flag.Var(&cfg.Env, "env", "Environment (dev|stage|prod)")
	flag.IntVar(&cfg.Port, "port", 8080, "Port")

	// Database configuration
	flag.StringVar(&cfg.Database.Driver, "database-driver", "sqlite", "database driver")
	flag.StringVar(&cfg.Database.Dsn, "database-dsn", "", "database dsn")

	flag.Parse()

	if cfg.Port < 0 || cfg.Port > 65535 {
		return nil, fmt.Errorf("port is not in valid range of 0-65535")
	}

	envs := []string{"sqlite"}
	if ok := slices.Contains(envs, cfg.Database.Driver); !ok {
		return nil, fmt.Errorf("env must be on of: %v", envs)
	}

	if cfg.Database.Dsn == "" {
		return nil, fmt.Errorf("database dsn cannot be empty")
	}

	return cfg, nil
}
