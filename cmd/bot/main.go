package main

import (
	"deadline-bot/internal/app"

	"flag"
	"log"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "configs/config.yaml", "Path to configuration file")
	flag.Parse()

	log.Println("Starting DeadlineBot...")

	application, err := app.NewApp(configPath)
	if err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
