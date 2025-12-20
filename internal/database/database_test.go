package database

import (
	"os"
	"testing"

	_ "modernc.org/sqlite"
)

func TestOpen_Success(t *testing.T) {
	db, err := Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("Open() failed: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		t.Errorf("Ping() failed: %v", err)
	}
}

func TestOpen_InvalidDriver(t *testing.T) {
	_, err := Open("invalid_driver", ":memory:")
	if err == nil {
		t.Error("expected error for invalid driver, got nil")
	}
}

func TestOpen_InvalidPath(t *testing.T) {
	_, err := Open("sqlite", "/invalid/path/that/does/not/exist/test.db")
	if err == nil {
		t.Error("expected error for invalid path, got nil")
	}
}

func TestVerify_ForeignKeyViolation(t *testing.T) {
	db, err := Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	schema, err := os.ReadFile("testdata/verify_test_schema.sql")
	if err != nil {
		t.Fatalf("failed to read schema file: %v", err)
	}

	if _, err := db.Exec(string(schema)); err != nil {
		t.Fatalf("failed to create tables: %v", err)
	}

	// Disable foreign keys temporarily to insert invalid data
	if _, err := db.Exec("PRAGMA foreign_keys=OFF;"); err != nil {
		t.Fatalf("failed to disable foreign keys: %v", err)
	}

	// Insert post with non-existent user_id
	if _, err := db.Exec("INSERT INTO posts (id, user_id) VALUES (1, 999);"); err != nil {
		t.Fatalf("failed to insert invalid data: %v", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys=ON;"); err != nil {
		t.Fatalf("failed to enable foreign keys: %v", err)
	}

	err = Verify(db)
	if err == nil {
		t.Error("Verify should have detected foreign key violation")
	} else if err.Error() != "foreign key violations detected" {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVerify_CorruptDatabase(t *testing.T) {
	// Todo: Implement test for corrupt database scenario using a corrupted database file.

	t.Skip("Requires manual database corruption")
}
