package utils

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestUUIDToString(t *testing.T) {
	uuidStr := "550e8400-e29b-41d4-a716-446655440000"
	var uuid pgtype.UUID
	if err := uuid.Scan(uuidStr); err != nil {
		t.Fatalf("Failed to scan uuid: %v", err)
	}

	got := UUIDToString(uuid)
	if got != uuidStr {
		t.Errorf("expected %s, got %s", uuidStr, got)
	}
}

func TestStringToUUID(t *testing.T) {
	uuidStr := "550e8400-e29b-41d4-a716-446655440000"
	uuid, err := StringToUUID(uuidStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !uuid.Valid {
		t.Error("expected valid uuid")
	}

	back := UUIDToString(uuid)
	if back != uuidStr {
		t.Errorf("expected %s, got %s", uuidStr, back)
	}
}
