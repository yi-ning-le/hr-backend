package service_test

import (
	"context"
	"testing"
	"time"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/service"
	"hr-backend/test/mocks"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestCreateCandidate(t *testing.T) {
	jobIDStr := "02020202-0202-0202-0202-020202020202"
	var jobIDUUID pgtype.UUID
	jobIDUUID.Scan(jobIDStr)

	mockRepo := &mocks.MockQuerier{
		CreateCandidateFunc: func(ctx context.Context, arg repository.CreateCandidateParams) (repository.Candidate, error) {
			return repository.Candidate{
				ID:           pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
				Name:         arg.Name,
				Email:        arg.Email,
				AppliedJobID: arg.AppliedJobID,
				Status:       arg.Status,
			}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			// GetCandidate is called after creation to fetch full details including joined job title
			return repository.GetCandidateRow{
				ID:              id,
				Name:            "John Doe",
				Email:           "john@example.com",
				AppliedJobID:    jobIDUUID,
				AppliedJobTitle: "Software Engineer",
				Status:          "new",
			}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)

	input := model.CandidateInput{
		Name:            "John Doe",
		Email:           "john@example.com",
		Phone:           "1234567890",
		ExperienceYears: 5,
		Education:       "BS CS",
		AppliedJobID:    jobIDStr,
		Channel:         "LinkedIn",
		ResumeURL:       "http://example.com/resume.pdf",
		AppliedAt:       time.Now(),
	}

	candidate, err := svc.CreateCandidate(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if candidate.Name != input.Name {
		t.Errorf("expected name %s, got %s", input.Name, candidate.Name)
	}
	if candidate.AppliedJobTitle != "Software Engineer" {
		t.Errorf("expected job title Software Engineer, got %s", candidate.AppliedJobTitle)
	}
}

func TestUpdateCandidate(t *testing.T) {
	cID := "00000000-0000-0000-0000-000000000001"
	jobIDStr := "02020202-0202-0202-0202-020202020202"
	var jobIDUUID pgtype.UUID
	jobIDUUID.Scan(jobIDStr)

	currentName := "Old Name"

	mockRepo := &mocks.MockQuerier{
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{
				ID:           id,
				Name:         currentName,
				AppliedJobID: jobIDUUID,
			}, nil
		},
		UpdateCandidateFunc: func(ctx context.Context, arg repository.UpdateCandidateParams) (repository.Candidate, error) {
			currentName = arg.Name
			return repository.Candidate{
				ID:   arg.ID,
				Name: arg.Name,
			}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	input := model.CandidateInput{
		Name:         "New Name",
		AppliedJobID: jobIDStr,
	}

	updated, err := svc.UpdateCandidate(context.Background(), cID, input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if updated.Name != "New Name" {
		t.Errorf("expected name New Name, got %s", updated.Name)
	}
}
