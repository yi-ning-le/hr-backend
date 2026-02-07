package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"hr-backend/internal/config"
	"hr-backend/internal/repository"
	"hr-backend/pkg/database"
)

func main() {
	username := flag.String("username", "", "Username to modify")
	setAdmin := flag.Bool("set-admin", false, "Set user as admin")
	removeAdmin := flag.Bool("remove-admin", false, "Remove admin from user")
	flag.Parse()

	if *username == "" {
		fmt.Println("Usage: go run cmd/admin/main.go -username <username> [-set-admin | -remove-admin]")
		os.Exit(1)
	}

	if !*setAdmin && !*removeAdmin {
		fmt.Println("Must specify either -set-admin or -remove-admin")
		os.Exit(1)
	}

	cfg := config.LoadConfig()
	db, err := database.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()
	queries := repository.New(db.Pool)

	// Find user by username
	user, err := queries.GetUserByUsername(ctx, *username)
	if err != nil {
		log.Fatalf("Failed to find user '%s': %v", *username, err)
	}

	// Update is_admin column directly
	var newValue bool
	if *setAdmin {
		newValue = true
	} else {
		newValue = false
	}

	_, err = db.Pool.Exec(ctx, "UPDATE users SET is_admin = $1 WHERE id = $2", newValue, user.ID)
	if err != nil {
		log.Fatalf("Failed to update user: %v", err)
	}

	if *setAdmin {
		fmt.Printf("User '%s' is now an Admin.\n", *username)
	} else {
		fmt.Printf("Admin removed from user '%s'.\n", *username)
	}
}
