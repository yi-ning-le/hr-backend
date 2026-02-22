package database

import "testing"

func TestMissingRequiredColumns(t *testing.T) {
	existing := map[string]struct{}{
		"id":         {},
		"user_id":    {},
		"event_type": {},
		"subject_id": {},
	}

	missing := missingRequiredColumns(existing)
	if len(missing) != 4 {
		t.Fatalf("expected 4 missing columns, got %d (%v)", len(missing), missing)
	}

	expected := map[string]bool{
		"subject_type": true,
		"context":      true,
		"read_at":      true,
		"created_at":   true,
	}
	for _, col := range missing {
		if !expected[col] {
			t.Fatalf("unexpected missing column: %s", col)
		}
		delete(expected, col)
	}
	if len(expected) != 0 {
		t.Fatalf("expected columns not found in missing list: %v", expected)
	}
}

func TestMissingRequiredColumns_NoneMissing(t *testing.T) {
	existing := make(map[string]struct{}, len(requiredNotificationColumns))
	for _, col := range requiredNotificationColumns {
		existing[col] = struct{}{}
	}

	missing := missingRequiredColumns(existing)
	if len(missing) != 0 {
		t.Fatalf("expected no missing columns, got %v", missing)
	}
}
