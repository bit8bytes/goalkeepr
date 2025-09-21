package main

import (
	"flag"
	"fmt"
	"os"
)

type config struct {
	env string
}

func main() {
	var cfg config
	if err := setup(&cfg); err != nil {
		os.Exit(1)
	}

	fmt.Printf("Hello from Goalkeepr.io! Ready for %s!\n", cfg.env)
}

func setup(cfg *config) error {
	flag.StringVar(&cfg.env, "env", "prod", "Environment (dev|stage|prod)")

	flag.Parse()

	return nil
}
