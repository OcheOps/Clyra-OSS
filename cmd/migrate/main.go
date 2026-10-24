package main

import (
	"log"

	"github.com/joho/godotenv"

	"github.com/oche/diagram/internal/config"
	"github.com/oche/diagram/internal/database"
)

func main() {
	_ = godotenv.Load()
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	pool, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("apply migrations: %v", err)
	}
	pool.Close()
	log.Print("database migrations are current")
}
