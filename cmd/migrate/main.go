package main

import (
	"context"
	"log"
	"os"
	"time"

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

	ctx := context.Background()

	// 3. Create migrations table if not exists
	_, err = db.Pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ DEFAULT NOW()
		);
	`)
	if err != nil {
		log.Fatalf("Failed to create migrations table: %v", err)
	}

	// 4. Get applied migrations
	rows, err := db.Pool.Query(ctx, "SELECT version FROM schema_migrations")
	if err != nil {
		log.Fatalf("Failed to query applied migrations: %v", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var v string
		if err := rows.Scan(&v); err != nil {
			log.Fatalf("Failed to scan migration version: %v", err)
		}
		applied[v] = true
	}

	// 5. Read migration directory
	entries, err := os.ReadDir("migrations")
	if err != nil {
		log.Fatalf("Failed to read migrations directory: %v", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && entry.Name()[len(entry.Name())-4:] == ".sql" {
			version := entry.Name()
			if !applied[version] {
				log.Printf("Applying migration: %s", version)
				content, err := os.ReadFile("migrations/" + version)
				if err != nil {
					log.Fatalf("Failed to read migration file %s: %v", version, err)
				}

				// Execute migration in transaction
				tx, err := db.Pool.Begin(ctx)
				if err != nil {
					log.Fatalf("Failed to start transaction for %s: %v", version, err)
				}

				_, err = tx.Exec(ctx, string(content))
				if err != nil {
					_ = tx.Rollback(ctx)
					log.Fatalf("Failed to execute migration %s: %v", version, err)
				}

				_, err = tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version)
				if err != nil {
					_ = tx.Rollback(ctx)
					log.Fatalf("Failed to record migration %s: %v", version, err)
				}

				if err := tx.Commit(ctx); err != nil {
					log.Fatalf("Failed to commit migration %s: %v", version, err)
				}
				log.Printf("Migration %s success", version)
			} else {
				log.Printf("Migration %s already applied, skipping", version)
			}
		}
	}

	log.Println("Migration process finished!")
	triggerAirReload()
}

func triggerAirReload() {
	now := time.Now().Format(time.RFC3339)
	if err := os.WriteFile("trigger/trigger.trigger", []byte(now), 0644); err != nil {
		log.Printf("Failed to write trigger file: %v", err)
		return
	}
	log.Println("Triggered Air hot reload")
}
