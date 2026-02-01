package main

import (
	"context"
	"log"
	"os"

	"hr-backend/internal/config"
	"hr-backend/pkg/database"
)

func main() {
	log.Println("Starting database migration...")

	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Connect to Database
	db, err := database.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// 3. Read Migration File
	migrationFile := "migrations/001_initial_schema.sql"
	content, err := os.ReadFile(migrationFile)
	if err != nil {
		log.Fatalf("Failed to read migration file %s: %v", migrationFile, err)
	}

	// 4. Execute Migration
	// We execute the raw SQL string. 
	// In a production app, you'd use a proper migration tool (like golang-migrate) 
	// to handle versioning and idempotent runs.
	_, err = db.Pool.Exec(context.Background(), string(content))
	if err != nil {
		log.Fatalf("Failed to execute migration: %v", err)
	}

	log.Println("Migration completed successfully!")
}
