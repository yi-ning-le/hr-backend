package repository

import (
	"strings"
	"testing"
)

func TestCreateInterviewUpsertDoesNotOverwriteCreatedByUserID(t *testing.T) {
	if strings.Contains(createInterview, "created_by_user_id = EXCLUDED.created_by_user_id") {
		t.Fatalf("createInterview upsert must not overwrite created_by_user_id")
	}
}
