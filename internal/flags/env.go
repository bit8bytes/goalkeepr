package flags

import (
	"fmt"
	"log/slog"
	"slices"
)

// Env represents the application environment.
type Env struct {
	slug string
}

func (e Env) LogValue() slog.Value {
	return slog.StringValue(e.slug)
}

// String returns the environment value.
func (e *Env) String() string {
	return e.slug
}

// Set validates and sets the environment value.
func (e *Env) Set(value string) error {
	envs := []string{"dev", "stage", "prod"}
	if !slices.Contains(envs, value) {
		return fmt.Errorf("env must be one of: %v", envs)
	}
	e.slug = value
	return nil
}

// IsDev returns true if environment is dev.
func (e *Env) IsDev() bool {
	return e.slug == "dev"
}

// SetEnv creates a new Env with the given value.
// Valid values are: dev, stage, prod.
func SetEnv(value string) Env {
	env := Env{}
	if err := env.Set(value); err != nil {
		panic(err)
	}
	return env
}
