package data

import (
	"testing"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name    string
		driver  string
		path    string
		wantErr bool
	}{
		{
			name:    "valid sqlite in-memory",
			driver:  "sqlite",
			path:    ":memory:",
			wantErr: false,
		},
		{
			name:    "unsupported driver",
			driver:  "mysql",
			path:    "any-path",
			wantErr: true,
		},
		{
			name:    "empty driver",
			driver:  "",
			path:    "any-path",
			wantErr: true,
		},
		{
			name:    "invalid sqlite path",
			driver:  "sqlite",
			path:    "/invalid/readonly/path/db.sqlite",
			wantErr: true,
		},
		{
			name:    "invalid postgres connection",
			driver:  "postgres",
			path:    "invalid-connection-string",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			New(test.driver, test.path)
		})
	}
}
