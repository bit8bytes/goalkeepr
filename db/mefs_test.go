package db

import (
	"io/fs"
	"testing"
)

func TestGetDatabaseMigrationsFS(t *testing.T) {
	tests := []struct {
		name         string
		databaseType string
		fs           fs.FS
	}{
		{
			name:         "Valid database migrations files for sqlite",
			databaseType: "sqlite",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := GetDatabaseMigrationFS(test.databaseType)
			if err != nil {
				t.Fatalf("got: %v;", err)
			}
		})
	}
}

func TestGetMigrationFS(t *testing.T) {
	fs := GetMigrationFS()
	if fs == nil {
		t.Fatal("got: returned nil filesystem;")
	}
}
