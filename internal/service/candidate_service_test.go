package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/service"
	"hr-backend/internal/utils"
	"hr-backend/test/mocks"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
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
		GetReviewerAssignmentFunc: func(ctx context.Context, arg repository.GetReviewerAssignmentParams) (repository.CandidateReviewer, error) {
			return repository.CandidateReviewer{}, pgx.ErrNoRows
		},
		SubmitReviewFunc: func(ctx context.Context, arg repository.SubmitReviewParams) (repository.SubmitReviewRow, error) {
			submitCalled = true
			return repository.SubmitReviewRow{}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, err := svc.SubmitReview(context.Background(), candidateIDStr, userIDStr, "suitable", "")
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
		"",
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
		"",
	)

	if !errors.Is(err, service.ErrCandidateNotFound) {
		t.Fatalf("expected ErrCandidateNotFound, got %v", err)
	}
}

func TestSubmitReview_PublishesReviewCompletedToRecruiter(t *testing.T) {
	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	reviewerUserIDStr := "00000000-0000-0000-0000-000000000002"
	reviewerEmployeeIDStr := "00000000-0000-0000-0000-000000000003"
	recruiterUserIDStr := "00000000-0000-0000-0000-000000000004"

	var candidateID, reviewerUserID, reviewerEmployeeID, recruiterUserID pgtype.UUID
	assert.NoError(t, candidateID.Scan(candidateIDStr))
	assert.NoError(t, reviewerUserID.Scan(reviewerUserIDStr))
	assert.NoError(t, reviewerEmployeeID.Scan(reviewerEmployeeIDStr))
	assert.NoError(t, recruiterUserID.Scan(recruiterUserIDStr))

	createNotificationCalled := false
	createdComments := make([]repository.CreateCandidateCommentParams, 0, 2)
	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{
				ID:        reviewerEmployeeID,
				UserID:    reviewerUserID,
				FirstName: "Alice",
				LastName:  "Lee",
			}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{
				ID:         candidateID,
				ReviewerID: reviewerEmployeeID,
			}, nil
		},
		GetReviewerAssignmentFunc: func(ctx context.Context, arg repository.GetReviewerAssignmentParams) (repository.CandidateReviewer, error) {
			return repository.CandidateReviewer{
				CandidateID:      candidateID,
				ReviewerID:       reviewerEmployeeID,
				AssignedByUserID: recruiterUserID,
			}, nil
		},
		SubmitReviewFunc: func(ctx context.Context, arg repository.SubmitReviewParams) (repository.SubmitReviewRow, error) {
			return repository.SubmitReviewRow{
				ID: candidateID,
			}, nil
		},
		CreateCandidateCommentFunc: func(ctx context.Context, arg repository.CreateCandidateCommentParams) (repository.CandidateComment, error) {
			createdComments = append(createdComments, arg)
			return repository.CandidateComment{}, nil
		},
		CreateNotificationFunc: func(ctx context.Context, arg repository.CreateNotificationParams) (repository.Notification, error) {
			createNotificationCalled = true
			assert.Equal(t, recruiterUserID, arg.UserID)
			assert.Equal(t, model.NotificationEventReviewCompleted, arg.EventType)
			assert.Equal(t, model.NotificationSubjectTypeCandidate, arg.SubjectType)
			assert.Equal(t, candidateID, arg.SubjectID)

			var payload map[string]string
			assert.NoError(t, json.Unmarshal(arg.Context, &payload))
			assert.Equal(t, candidateIDStr, payload["candidateId"])
			assert.Equal(t, "suitable", payload["reviewStatus"])
			assert.Equal(t, "Alice Lee", payload["reviewerName"])
			return repository.Notification{}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, err := svc.SubmitReview(
		context.Background(),
		candidateIDStr,
		reviewerUserIDStr,
		"suitable",
		"Strong fit",
	)
	assert.NoError(t, err)
	assert.Len(t, createdComments, 2)
	assert.Equal(t, "Strong fit", createdComments[0].Content)
	assert.Equal(t, "normal", createdComments[0].CommentType)
	assert.Equal(t, "suitable", createdComments[1].Content)
	assert.Equal(t, "review_suitable", createdComments[1].CommentType)
	assert.True(t, createNotificationCalled)
}

func TestAssignReviewer_ReplacesActiveAssignment(t *testing.T) {
	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	reviewerIDStr := "00000000-0000-0000-0000-000000000002"
	assignedByUserIDStr := "00000000-0000-0000-0000-00000000000a"

	var candidateID, reviewerID, assignedByUserID pgtype.UUID
	if err := candidateID.Scan(candidateIDStr); err != nil {
		t.Fatalf("failed to scan candidate id: %v", err)
	}
	if err := reviewerID.Scan(reviewerIDStr); err != nil {
		t.Fatalf("failed to scan reviewer id: %v", err)
	}
	if err := assignedByUserID.Scan(assignedByUserIDStr); err != nil {
		t.Fatalf("failed to scan assigned-by user id: %v", err)
	}

	callOrder := make([]string, 0, 3)
	mockUserID := pgtype.UUID{Bytes: [16]byte{9, 9, 9}, Valid: true}
	mockRepo := &mocks.MockQuerier{
		GetEmployeeFunc: func(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
			if id != reviewerID {
				return repository.Employee{}, pgx.ErrNoRows
			}
			return repository.Employee{ID: reviewerID, UserID: mockUserID}, nil
		},
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
			if arg.AssignedByUserID != assignedByUserID {
				t.Fatalf("expected assigned_by_user_id %v, got %v", assignedByUserID, arg.AssignedByUserID)
			}
			return repository.CandidateReviewer{}, nil
		},
		CreateNotificationFunc: func(ctx context.Context, arg repository.CreateNotificationParams) (repository.Notification, error) {
			callOrder = append(callOrder, "create_notification")
			if arg.UserID != mockUserID {
				t.Fatalf("expected notification user id %v, got %v", mockUserID, arg.UserID)
			}
			assert.Equal(t, model.NotificationEventCandidateReviewerAssigned, arg.EventType)
			assert.Equal(t, model.NotificationSubjectTypeCandidate, arg.SubjectType)
			assert.Equal(t, candidateID, arg.SubjectID)
			var contextPayload map[string]string
			assert.NoError(t, json.Unmarshal(arg.Context, &contextPayload))
			assert.Equal(t, utils.UUIDToString(candidateID), contextPayload["candidateId"])
			return repository.Notification{}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, err := svc.AssignReviewer(context.Background(), candidateIDStr, reviewerIDStr, assignedByUserIDStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(callOrder) != 3 || callOrder[0] != "remove_old" || callOrder[1] != "insert_new" || callOrder[2] != "create_notification" {
		t.Fatalf("unexpected call order: %v", callOrder)
	}
}

func TestSubmitReview_CreateCommentFailure_ReturnsErrorAndSkipsNotification(t *testing.T) {
	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	reviewerUserIDStr := "00000000-0000-0000-0000-000000000002"
	reviewerEmployeeIDStr := "00000000-0000-0000-0000-000000000003"
	recruiterUserIDStr := "00000000-0000-0000-0000-000000000004"

	var candidateID, reviewerUserID, reviewerEmployeeID, recruiterUserID pgtype.UUID
	assert.NoError(t, candidateID.Scan(candidateIDStr))
	assert.NoError(t, reviewerUserID.Scan(reviewerUserIDStr))
	assert.NoError(t, reviewerEmployeeID.Scan(reviewerEmployeeIDStr))
	assert.NoError(t, recruiterUserID.Scan(recruiterUserIDStr))

	createNotificationCalled := false
	expectedErr := errors.New("insert comment failed")
	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{
				ID:        reviewerEmployeeID,
				UserID:    reviewerUserID,
				FirstName: "Alice",
				LastName:  "Lee",
			}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{
				ID:         candidateID,
				ReviewerID: reviewerEmployeeID,
			}, nil
		},
		GetReviewerAssignmentFunc: func(ctx context.Context, arg repository.GetReviewerAssignmentParams) (repository.CandidateReviewer, error) {
			return repository.CandidateReviewer{
				CandidateID:      candidateID,
				ReviewerID:       reviewerEmployeeID,
				AssignedByUserID: recruiterUserID,
			}, nil
		},
		SubmitReviewFunc: func(ctx context.Context, arg repository.SubmitReviewParams) (repository.SubmitReviewRow, error) {
			return repository.SubmitReviewRow{ID: candidateID}, nil
		},
		CreateCandidateCommentFunc: func(ctx context.Context, arg repository.CreateCandidateCommentParams) (repository.CandidateComment, error) {
			return repository.CandidateComment{}, expectedErr
		},
		CreateNotificationFunc: func(ctx context.Context, arg repository.CreateNotificationParams) (repository.Notification, error) {
			createNotificationCalled = true
			return repository.Notification{}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, err := svc.SubmitReview(
		context.Background(),
		candidateIDStr,
		reviewerUserIDStr,
		"suitable",
		"Strong fit",
	)
	assert.ErrorIs(t, err, expectedErr)
	assert.False(t, createNotificationCalled)
}

func TestSubmitReview_PendingSkipsDecisionCommentAndNotification(t *testing.T) {
	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	reviewerUserIDStr := "00000000-0000-0000-0000-000000000002"
	reviewerEmployeeIDStr := "00000000-0000-0000-0000-000000000003"
	recruiterUserIDStr := "00000000-0000-0000-0000-000000000004"

	var candidateID, reviewerUserID, reviewerEmployeeID, recruiterUserID pgtype.UUID
	assert.NoError(t, candidateID.Scan(candidateIDStr))
	assert.NoError(t, reviewerUserID.Scan(reviewerUserIDStr))
	assert.NoError(t, reviewerEmployeeID.Scan(reviewerEmployeeIDStr))
	assert.NoError(t, recruiterUserID.Scan(recruiterUserIDStr))

	createNotificationCalled := false
	createdComments := make([]repository.CreateCandidateCommentParams, 0, 2)
	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{
				ID:        reviewerEmployeeID,
				UserID:    reviewerUserID,
				FirstName: "Alice",
				LastName:  "Lee",
			}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{
				ID:         candidateID,
				ReviewerID: reviewerEmployeeID,
			}, nil
		},
		GetReviewerAssignmentFunc: func(ctx context.Context, arg repository.GetReviewerAssignmentParams) (repository.CandidateReviewer, error) {
			return repository.CandidateReviewer{
				CandidateID:      candidateID,
				ReviewerID:       reviewerEmployeeID,
				AssignedByUserID: recruiterUserID,
			}, nil
		},
		SubmitReviewFunc: func(ctx context.Context, arg repository.SubmitReviewParams) (repository.SubmitReviewRow, error) {
			return repository.SubmitReviewRow{ID: candidateID}, nil
		},
		CreateCandidateCommentFunc: func(ctx context.Context, arg repository.CreateCandidateCommentParams) (repository.CandidateComment, error) {
			createdComments = append(createdComments, arg)
			return repository.CandidateComment{}, nil
		},
		CreateNotificationFunc: func(ctx context.Context, arg repository.CreateNotificationParams) (repository.Notification, error) {
			createNotificationCalled = true
			return repository.Notification{}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, err := svc.SubmitReview(
		context.Background(),
		candidateIDStr,
		reviewerUserIDStr,
		"pending",
		"Need follow-up",
	)
	assert.NoError(t, err)
	assert.Len(t, createdComments, 1)
	assert.Equal(t, "normal", createdComments[0].CommentType)
	assert.Equal(t, "Need follow-up", createdComments[0].Content)
	assert.False(t, createNotificationCalled)
}
