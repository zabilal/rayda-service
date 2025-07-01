//go:build ignore

package main

import (
	"context"
	"flag"
	"log"

	"github.com/rayda/rayda-service/internal/config"
	"github.com/rayda/rayda-service/internal/database"
)

func main() {
	// Parse command line flags
	action := flag.String("action", "up", "Migration action: up, down, or reset")
	// envFile := flag.String("env", ".env", "Path to .env file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to the database
	if err := database.Connect(cfg); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Perform the requested action
	switch *action {
	case "up":
		log.Println("Running database migrations...")
		if err := database.Migrate(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Database migrations completed successfully")

	case "reset":
		log.Println("Resetting database...")
		// Drop all tables and re-run migrations
		db, err := database.DB.DB()
		if err != nil {
			log.Fatalf("Failed to get database instance: %v", err)
		}

		// Drop all tables
		_, err = db.ExecContext(context.Background(), `
			DROP SCHEMA public CASCADE;
			CREATE SCHEMA public;
			GRANT ALL ON SCHEMA public TO postgres;
			GRANT ALL ON SCHEMA public TO public;
		`)
		if err != nil {
			log.Fatalf("Failed to reset database: %v", err)
		}

		// Re-run migrations
		if err := database.Migrate(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
		log.Println("Database reset and migrations completed successfully")

	default:
		log.Fatalf("Unknown action: %s", *action)
	}
}
