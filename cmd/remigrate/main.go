package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"

	"hr-backend/internal/config"
	"hr-backend/pkg/database"
)

var migrationsDir = "migrations"

func main() {
	flag.Parse()
	versions := flag.Args()

	if len(versions) == 0 {
		log.Fatal("请指定要重新运行的迁移文件，例如: go run cmd/remigrate/main.go 005_recruitment_schema.sql")
	}

	cfg := config.LoadConfig()
	db, err := database.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	for _, version := range versions {
		if err := rerunMigration(ctx, db, version); err != nil {
			log.Fatalf("Failed to rerun migration %s: %v", version, err)
		}
	}

	log.Println("All migrations re-applied successfully!")
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

func rerunMigration(ctx context.Context, db *database.Database, version string) error {
	log.Printf("Processing migration: %s", version)

	content, err := os.ReadFile(migrationsDir + "/" + version)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	tables := extractTables(string(content))
	if len(tables) == 0 {
		log.Printf("No tables found in %s, skipping drop", version)
	} else {
		log.Printf("Tables to drop: %v", tables)
		for _, table := range tables {
			dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)
			if _, err := db.Pool.Exec(ctx, dropSQL); err != nil {
				return fmt.Errorf("failed to drop table %s: %w", table, err)
			}
			log.Printf("Dropped table: %s", table)
		}
	}

	_, err = db.Pool.Exec(ctx, "DELETE FROM schema_migrations WHERE version = $1", version)
	if err != nil {
		return fmt.Errorf("failed to delete migration record: %w", err)
	}
	log.Printf("Removed migration record: %s", version)

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			log.Printf("rollback error: %v", err)
		}
	}()

	_, err = tx.Exec(ctx, string(content))
	if err != nil {
		return fmt.Errorf("failed to execute migration: %w", err)
	}

	_, err = tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit migration: %w", err)
	}

	log.Printf("Migration %s re-applied successfully", version)
	return nil
}

func extractTables(sql string) []string {
	tableSet := make(map[string]bool)

	createTableRegex := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`)
	alterTableRegex := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+(?:ONLY\s+)?(\w+)`)
	dropTableRegex := regexp.MustCompile(`(?i)DROP\s+TABLE\s+(?:IF\s+EXISTS\s+)?(\w+)`)

	patterns := []*regexp.Regexp{createTableRegex, alterTableRegex, dropTableRegex}
	for _, pattern := range patterns {
		matches := pattern.FindAllStringSubmatch(sql, -1)
		for _, match := range matches {
			if len(match) > 1 {
				tableSet[strings.ToLower(match[1])] = true
			}
		}
	}

	tables := make([]string, 0, len(tableSet))
	for table := range tableSet {
		tables = append(tables, table)
	}
	return tables
}
