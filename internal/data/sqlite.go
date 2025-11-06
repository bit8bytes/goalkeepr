package data

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"time"

	"github.com/pressly/goose/v3"
)

// SQLiteProvider implements Provider interface for SQLite databases
type SQLiteProvider struct{}

// Open establishes a connection to a SQLite database
func (p *SQLiteProvider) Open(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("could not open db connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		db.Close()
		return nil, err
	}

	// Readers and writers don't block each other.
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, fmt.Errorf("could not set journal mode: %w", err)
	}

	_, err = db.Exec("PRAGMA synchronous=NORMAL")
	if err != nil {
		return nil, fmt.Errorf("could not set synchronous mode: %w", err)
	}

	// prevent panic under write pressure
	_, err = db.Exec("PRAGMA busy_timeout = 5000;")
	if err != nil {
		return nil, fmt.Errorf("could not set busy timeout: %w", err)
	}

	_, err = db.Exec("PRAGMA foreign_keys=ON;")
	if err != nil {
		return nil, fmt.Errorf("could not enable foreign keys: %w", err)
	}

	return db, nil
}

func (p *SQLiteProvider) GetDSN(path string) string {
	return fmt.Sprintf("%s", path)
}

func (p *SQLiteProvider) AutoMigrate(migrationsFS fs.FS, path string, db *sql.DB) (*int64, error) {
	goose.SetBaseFS(migrationsFS)
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("sqlite"); err != nil {
		return nil, fmt.Errorf("could not set goose dialect: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return nil, fmt.Errorf("goose migration failed: %w", err)
	}

	version, err := goose.GetDBVersion(db)
	if err != nil {
		return nil, fmt.Errorf("could not get version: %w", err)
	}

	return &version, nil
}
