package flags

import (
	"testing"
)

func TestEnv_Set_ValidValues(t *testing.T) {
	tests := []string{"dev", "stage", "prod"}

	for _, value := range tests {
		t.Run(value, func(t *testing.T) {
			env := &Env{}
			if err := env.Set(value); err != nil {
				t.Errorf("Set(%q) failed: %v", value, err)
			}
			if env.String() != value {
				t.Errorf("expected %q, got %q", value, env.String())
			}
		})
	}
}

func TestEnv_Set_InvalidValue(t *testing.T) {
	env := &Env{}
	if err := env.Set("invalid"); err == nil {
		t.Error("expected error for invalid environment, got nil")
	}
}

func TestNewEnv_ValidValue(t *testing.T) {
	env := SetEnv("dev")
	if env.String() != "dev" {
		t.Errorf("expected 'dev', got %q", env.String())
	}
}

func TestNewEnv_InvalidValue_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid value, got none")
		}
	}()
	SetEnv("invalid")
}
