package data

import (
	"database/sql"
	"fmt"
	"io/fs"
	"strings"

	"github.com/bit8bytes/goalkeepr/db"
)

var SupporedDrivers = []string{"sqlite"}

// Database provider interface
type DatabaseProvider interface {
	Open(path string) (*sql.DB, error)
	GetDSN(path string) string
	AutoMigrate(migrationsFS fs.FS, path string, db *sql.DB) (version *int64, err error)
}

// Database holds the configuration for the database connection
type Database struct {
	Driver     string
	Path       string
	Migrations string
	DB         *sql.DB
	provider   DatabaseProvider
}

func New(driver, path string) (*Database, error) {
	db := &Database{
		Driver: driver,
		Path:   path,
	}

	switch driver {
	case "sqlite":
		db.provider = &SQLiteProvider{}
	default:
		return nil, fmt.Errorf(
			"unsupported database driver: %s. Supported drivers: %s",
			driver,
			strings.Join(SupporedDrivers, ","))
	}

	conn, err := db.provider.Open(path)
	if err != nil {
		return nil, err
	}
	db.DB = conn

	return db, nil
}

// AutoMigrate creates necessary database tables using embedded migrations and returns the version
func (d *Database) AutoMigrate() (version *int64, err error) {
	// Get the migrations for the current database type
	migrationsFS, err := db.GetDatabaseMigrationFS(d.Driver)
	if err != nil {
		return nil, fmt.Errorf("failed to get migrations for %s: %w", d.Driver, err)
	}

	return d.provider.AutoMigrate(migrationsFS, d.Path, d.DB)
}
