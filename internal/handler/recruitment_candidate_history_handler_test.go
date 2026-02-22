package handler

import (
	"testing"
	"time"

	"hr-backend/internal/repository"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestMapPastReviewedCandidateRows_ReviewedAtMapping(t *testing.T) {
	var candidateID pgtype.UUID
	err := candidateID.Scan("11111111-1111-1111-1111-111111111111")
	assert.NoError(t, err)

	var jobID pgtype.UUID
	err = jobID.Scan("22222222-2222-2222-2222-222222222222")
	assert.NoError(t, err)

	reviewedTime := time.Date(2025, 1, 10, 14, 30, 0, 0, time.UTC)
	appliedTime := time.Date(2025, 1, 5, 9, 0, 0, 0, time.UTC)

	rows := []repository.GetPastReviewedCandidatesRow{
		{
			ID:              candidateID,
			Name:            "Alice",
			Email:           "alice@example.com",
			Phone:           "10000000000",
			ExperienceYears: 5,
			Education:       "BS",
			AppliedJobID:    jobID,
			AppliedJobTitle: "Engineer",
			Channel:         "LinkedIn",
			ResumeUrl:       "https://example.com/a.pdf",
			Status:          "interview",
			ReviewStatus:    "suitable",
			AppliedAt:       pgtype.Timestamptz{Time: appliedTime, Valid: true},
			ReviewedAt:      pgtype.Timestamptz{Time: reviewedTime, Valid: true},
		},
		{
			ID:              candidateID,
			Name:            "Bob",
			Email:           "bob@example.com",
			Phone:           "10000000001",
			ExperienceYears: 3,
			Education:       "MS",
			AppliedJobID:    jobID,
			AppliedJobTitle: "Designer",
			Channel:         "Referral",
			ResumeUrl:       "https://example.com/b.pdf",
			Status:          "new",
			ReviewStatus:    "unsuitable",
			AppliedAt:       pgtype.Timestamptz{Time: appliedTime, Valid: true},
			ReviewedAt:      pgtype.Timestamptz{Valid: false},
		},
	}

	items := mapPastReviewedCandidateRows(rows)

	assert.Len(t, items, 2)
	assert.NotNil(t, items[0].ReviewedAt)
	assert.Equal(t, reviewedTime, *items[0].ReviewedAt)
	assert.Nil(t, items[1].ReviewedAt)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", items[0].ID)
	assert.Equal(t, "22222222-2222-2222-2222-222222222222", items[0].AppliedJobID)
}
