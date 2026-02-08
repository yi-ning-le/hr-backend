package main

import (
	"context"
	"fmt"
	"log"

	"hr-backend/internal/config"
	"hr-backend/pkg/database"

	"github.com/jackc/pgx/v5/pgtype"
)

func main() {
	// Load Config
	cfg := config.LoadConfig()

	// Database Connection
	db, err := database.NewDatabase(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	ctx := context.Background()

	// 1. Find User by Username 'admin' (since that's the likely user)
	var userID pgtype.UUID
	var username string
	err = db.Pool.QueryRow(ctx, "SELECT id, username FROM users WHERE username='admin'").Scan(&userID, &username)
	if err != nil {
		fmt.Println("User 'admin' NOT FOUND. Listing all users:")
		rows, _ := db.Pool.Query(ctx, "SELECT id, username FROM users LIMIT 10")
		for rows.Next() {
			var uid pgtype.UUID
			var uname string
			rows.Scan(&uid, &uname)
			fmt.Printf(" - %s (%v)\n", uname, uid)
		}
		return
	}
	fmt.Printf("Found User 'admin': %v\n", userID)

	// 2. Find Linked Employee
	var empID pgtype.UUID
	var empName string
	err = db.Pool.QueryRow(ctx, "SELECT id, first_name FROM employees WHERE user_id=$1", userID).Scan(&empID, &empName)
	if err != nil {
		fmt.Printf("User 'admin' has NO LINKED EMPLOYEE. Creating one...\n")
		// Create employee
		var newEmpID pgtype.UUID
		err = db.Pool.QueryRow(ctx, `
            INSERT INTO employees (first_name, last_name, email, department, position, status, employment_type, join_date, user_id)
            VALUES ('Admin', 'User', 'admin@example.com', 'Management', 'Administrator', 'Active', 'FullTime', NOW(), $1)
            RETURNING id
        `, userID).Scan(&newEmpID)
		if err != nil {
			log.Fatalf("Failed to create employee: %v", err)
		}
		empID = newEmpID
		fmt.Printf("Created Employee %v for 'admin'.\n", empID)
	} else {
		fmt.Printf("User 'admin' is linked to Employee %v (%s)\n", empID, empName)
	}

	// 3. Assign Recruiter Role
	_, err = db.Pool.Exec(ctx, "INSERT INTO recruitment_roles (employee_id, role_type) VALUES ($1, 'RECRUITER') ON CONFLICT DO NOTHING", empID)
	if err != nil {
		fmt.Printf("Failed to assign Recruiter role: %v\n", err)
	} else {
		fmt.Printf("SUCCESS: Assigned 'RECRUITER' role to 'admin' (Employee %v).\n", empID)
	}
}
