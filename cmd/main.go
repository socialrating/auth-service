package main

import (
	"context"
	"flag"
	"log"

	"github.com/socialrating/auth-service/config"
	"github.com/socialrating/auth-service/internal/application"
)

func main() {
	cfgPath := flag.String("config", "path/to/config.yaml", "Path to the configuration file")
	flag.Parse()

	// через flag передавать config при запуске сервиса
	ctx := context.Background()

	cfg, err := config.LoadConfig(*cfgPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	if err := application.Run(ctx, cfg); err != nil {
		log.Fatalf("failed to run application: %v", err)
	}
}
