package main

import (
	"flag"
	"os"
	"testing"
)

func TestSetup(t *testing.T) {
	var tests = []struct {
		name          string
		expectedValue string
		wantErr       error
	}{
		{
			name:          "Valid dev env",
			expectedValue: "dev",
			wantErr:       nil,
		},
		{
			name:          "Valid stage env",
			expectedValue: "stage",
			wantErr:       nil,
		},
		{
			name:          "Valid prod env",
			expectedValue: "prod",
			wantErr:       nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError) // Reset the flags each test
			os.Args = []string{"-env", test.expectedValue}

			var cfg config
			err := setup(&cfg)
			if err != test.wantErr {
				t.Errorf("git: %v; want %v;", err, test.wantErr)
			}
		})
	}
}
