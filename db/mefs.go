//go:build !dev

package db

import (
	"embed"
	"io/fs"
	"path/filepath"
)

//go:embed "migrations"
var migrationFiles embed.FS

// GetMigrationFS returns the full migrations filesystem
func GetMigrationFS() fs.FS {
	return migrationFiles
}

// GetDatabaseMigrationFS returns migrations for a specific database provider
func GetDatabaseMigrationFS(dbType string) (fs.FS, error) {
	migrationPath := filepath.Join("migrations", dbType)

	subFS, err := fs.Sub(migrationFiles, migrationPath)
	if err != nil {
		return nil, err
	}
	return subFS, nil
}
