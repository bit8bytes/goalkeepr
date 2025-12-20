package database

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"time"

	"github.com/pressly/goose/v3"
)

func Open(driver, path string) (*sql.DB, error) {
	db, err := sql.Open(driver, path)
	if err != nil {
		return nil, fmt.Errorf("could not open db connection: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	err = db.PingContext(ctx)
	if err != nil {
		_ = db.Close() // Ignore close error, ping error takes precedence
		return nil, err
	}

	// Readers and writers don't block each other.
	_, err = db.Exec("PRAGMA journal_mode=WAL;")
	if err != nil {
		return nil, fmt.Errorf("could not set journal mode: %w", err)
	}

	// The trusted_schema setting is a per-connection
	// boolean that determines whether or not SQL functions
	// and virtual tables that have not been security audited
	// are allowed to be run by views, triggers, or
	// in expressions
	_, err = db.Exec("PRAGMA trusted_schema=OFF;")
	if err != nil {
		return nil, fmt.Errorf("could not set trusted schema: %w", err)
	}

	// This ensures that an operating system crash or power failure will not corrupt the database.
	// This can be changed to NORMAL for better performance when backup/rollback is in place.
	_, err = db.Exec("PRAGMA synchronous=FULL")
	if err != nil {
		return nil, fmt.Errorf("could not set synchronous mode: %w", err)
	}

	// Prevent panic under write pressure
	// This routine sets a busy handler that sleeps for a specified amount of time when a table is locked
	_, err = db.Exec("PRAGMA busy_timeout = 1000;")
	if err != nil {
		return nil, fmt.Errorf("could not set busy timeout: %w", err)
	}

	_, err = db.Exec("PRAGMA foreign_keys=ON;")
	if err != nil {
		return nil, fmt.Errorf("could not enable foreign keys: %w", err)
	}

	// When secure_delete is set to "fast", SQLite will
	// overwrite deleted content with zeros only if doing
	// so does not increase the amount of I/O
	_, err = db.Exec("PRAGMA secure_delete = FAST;")
	if err != nil {
		return nil, fmt.Errorf("could not set secure delete mode: %w", err)
	}

	// Database corruption is detected earlier and is less likely to "spread".
	// However, there is a small performance hit for doing the extra checks and so cell size checking
	_, err = db.Exec("PRAGMA cell_size_check=ON;")
	if err != nil {
		return nil, fmt.Errorf("could not enable cell size check: %w", err)
	}

	// SQLite pool settings
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetConnMaxIdleTime(5 * time.Minute)

	return db, nil
}

func Migrate(db *sql.DB, migrations fs.FS) (int64, error) {
	goose.SetBaseFS(migrations)
	goose.SetLogger(goose.NopLogger())

	if err := goose.SetDialect("sqlite"); err != nil {
		return 0, fmt.Errorf("could not set goose dialect: %w", err)
	}

	if err := goose.Up(db, "db/migrations"); err != nil {
		return 0, fmt.Errorf("goose migration failed: %w", err)
	}

	version, err := goose.GetDBVersion(db)
	if err != nil {
		return 0, fmt.Errorf("could not get version: %w", err)
	}

	return version, nil
}

func Verify(db *sql.DB) error {
	rows, err := db.Query("PRAGMA foreign_key_check;")
	if err != nil {
		return fmt.Errorf("could not run foreign key check: %w", err)
	}
	defer rows.Close()

	if rows.Next() {
		return fmt.Errorf("foreign key violations detected")
	}

	rows, err = db.Query("PRAGMA integrity_check;")
	if err != nil {
		return fmt.Errorf("could not run integrity check: %w", err)
	}
	defer rows.Close()

	var result string
	if rows.Next() {
		if err := rows.Scan(&result); err != nil {
			return fmt.Errorf("could not scan integrity check result: %w", err)
		}
		if result != "ok" {
			return fmt.Errorf("integrity check failed: %s", result)
		}
	}

	return nil
}

func Optimize(db *sql.DB) error {
	// Applications that use long-lived database connections should run "PRAGMA optimize=0x10002;"
	// when the connection is first opened, and then also run "PRAGMA optimize;" periodically,
	// perhaps once per day or once per hour.
	_, err := db.Exec("PRAGMA optimize=0x10002;")
	if err != nil {
		return fmt.Errorf("could not optimize database: %w", err)
	}
	return nil
}
