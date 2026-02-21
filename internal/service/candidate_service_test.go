package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/service"
	"hr-backend/test/mocks"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestCreateCandidate(t *testing.T) {
	jobIDStr := "02020202-0202-0202-0202-020202020202"
	var jobIDUUID pgtype.UUID
	if err := jobIDUUID.Scan(jobIDStr); err != nil {
		t.Fatalf("failed to scan job id: %v", err)
	}

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
	if err := jobIDUUID.Scan(jobIDStr); err != nil {
		t.Fatalf("failed to scan job id: %v", err)
	}

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

func TestSubmitReview_OnlyAssignedReviewer(t *testing.T) {
	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	userIDStr := "00000000-0000-0000-0000-000000000002"

	var candidateID pgtype.UUID
	if err := candidateID.Scan(candidateIDStr); err != nil {
		t.Fatalf("failed to scan candidate id: %v", err)
	}

	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		t.Fatalf("failed to scan user id: %v", err)
	}

	var actorEmployeeID pgtype.UUID
	if err := actorEmployeeID.Scan("00000000-0000-0000-0000-000000000003"); err != nil {
		t.Fatalf("failed to scan actor employee id: %v", err)
	}

	var assignedReviewerID pgtype.UUID
	if err := assignedReviewerID.Scan("00000000-0000-0000-0000-000000000004"); err != nil {
		t.Fatalf("failed to scan assigned reviewer id: %v", err)
	}

	submitCalled := false
	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{ID: actorEmployeeID, UserID: userID}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{
				ID:         candidateID,
				ReviewerID: assignedReviewerID,
			}, nil
		},
		SubmitReviewFunc: func(ctx context.Context, arg repository.SubmitReviewParams) (repository.SubmitReviewRow, error) {
			submitCalled = true
			return repository.SubmitReviewRow{}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, err := svc.SubmitReview(context.Background(), candidateIDStr, userIDStr, "suitable",)
	if !errors.Is(err, service.ErrReviewPermissionDenied) {
		t.Fatalf("expected ErrReviewPermissionDenied, got %v", err)
	}
	if submitCalled {
		t.Fatalf("expected SubmitReview not to be called when reviewer is not assigned")
	}
}

func TestSubmitReview_ReviewerProfileNotFound(t *testing.T) {
	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{}, pgx.ErrNoRows
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, err := svc.SubmitReview(
		context.Background(),
		"00000000-0000-0000-0000-000000000001",
		"00000000-0000-0000-0000-000000000002",
		"suitable",
		
	)

	if !errors.Is(err, service.ErrReviewerProfileNotFound) {
		t.Fatalf("expected ErrReviewerProfileNotFound, got %v", err)
	}
}

func TestSubmitReview_CandidateNotFound(t *testing.T) {
	var employeeID pgtype.UUID
	if err := employeeID.Scan("00000000-0000-0000-0000-000000000003"); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{ID: employeeID, UserID: id}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{}, pgx.ErrNoRows
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, err := svc.SubmitReview(
		context.Background(),
		"00000000-0000-0000-0000-000000000001",
		"00000000-0000-0000-0000-000000000002",
		"suitable",
		
	)

	if !errors.Is(err, service.ErrCandidateNotFound) {
		t.Fatalf("expected ErrCandidateNotFound, got %v", err)
	}
}

func TestAssignReviewer_ReplacesActiveAssignment(t *testing.T) {
	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	reviewerIDStr := "00000000-0000-0000-0000-000000000002"

	var candidateID, reviewerID pgtype.UUID
	if err := candidateID.Scan(candidateIDStr); err != nil {
		t.Fatalf("failed to scan candidate id: %v", err)
	}
	if err := reviewerID.Scan(reviewerIDStr); err != nil {
		t.Fatalf("failed to scan reviewer id: %v", err)
	}

	callOrder := make([]string, 0, 2)
	mockRepo := &mocks.MockQuerier{
		AssignReviewerFunc: func(ctx context.Context, arg repository.AssignReviewerParams) (repository.AssignReviewerRow, error) {
			return repository.AssignReviewerRow{
				ID:         arg.ID,
				ReviewerID: arg.ReviewerID,
			}, nil
		},
		UpdateCandidateReviewerRemovedAtFunc: func(ctx context.Context, id pgtype.UUID) error {
			callOrder = append(callOrder, "remove_old")
			if id != candidateID {
				t.Fatalf("expected candidate id %v, got %v", candidateID, id)
			}
			return nil
		},
		InsertCandidateReviewerFunc: func(ctx context.Context, arg repository.InsertCandidateReviewerParams) (repository.CandidateReviewer, error) {
			callOrder = append(callOrder, "insert_new")
			if arg.CandidateID != candidateID {
				t.Fatalf("expected candidate id %v, got %v", candidateID, arg.CandidateID)
			}
			if arg.ReviewerID != reviewerID {
				t.Fatalf("expected reviewer id %v, got %v", reviewerID, arg.ReviewerID)
			}
			return repository.CandidateReviewer{}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, err := svc.AssignReviewer(context.Background(), candidateIDStr, reviewerIDStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(callOrder) != 2 || callOrder[0] != "remove_old" || callOrder[1] != "insert_new" {
		t.Fatalf("unexpected call order: %v", callOrder)
	}
}
