package database

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

var requiredNotificationColumns = []string{
	"id",
	"user_id",
	"event_type",
	"subject_type",
	"subject_id",
	"context",
	"read_at",
	"created_at",
}

func EnsureNotificationSchema(ctx context.Context, db *Database) error {
	rows, err := db.Pool.Query(ctx, `
		SELECT column_name
		FROM information_schema.columns
		WHERE table_schema = current_schema()
		  AND table_name = 'notifications'
	`)
	if err != nil {
		return fmt.Errorf("query notifications schema failed: %w", err)
	}
	defer rows.Close()

	existing := make(map[string]struct{}, len(requiredNotificationColumns))
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("scan notifications schema failed: %w", err)
		}
		existing[name] = struct{}{}
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate notifications schema failed: %w", err)
	}

	missing := missingRequiredColumns(existing)
	if len(missing) > 0 {
		return fmt.Errorf("notifications schema mismatch, missing columns: %s", strings.Join(missing, ", "))
	}

	return nil
}

func missingRequiredColumns(existing map[string]struct{}) []string {
	missing := make([]string, 0, len(requiredNotificationColumns))
	for _, col := range requiredNotificationColumns {
		if _, ok := existing[col]; !ok {
			missing = append(missing, col)
		}
	}
	sort.Strings(missing)
	return missing
}
