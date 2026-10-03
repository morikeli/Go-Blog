package main

import (
	"log"

	"github.com/morikeli/golangrestapi/internal/config"
)

func main() {
	// load project config
	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatalf("Failed to load project configuration file: %v", err)
	}

	application, err := NewApp(cfg)
	if err != nil {
		log.Fatalf("Failed to run application: %v", err)
	}

	if err := application.Run(cfg.ServerPort); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
