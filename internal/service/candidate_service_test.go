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

	input := model.CandidateCreateInput{
		Name:            "John Doe",
		Email:           "john@example.com",
		Phone:           "1234567890",
		ExperienceYears: 5,
		Education:       "BS CS",
		AppliedJobID:    jobIDStr,
		Channel:         "LinkedIn",
		AppliedAt:       time.Now(),
	}

	candidate, err := svc.CreateCandidate(context.Background(), input, "http://example.com/resume.pdf")
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
				Name:       "John Doe",
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
			assert.Equal(t, "John Doe", payload["candidateName"])
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

func TestRevertReviewer_InvalidCandidateID_ReturnsErrInvalidCandidateID(t *testing.T) {
	svc := service.NewCandidateService(&mocks.MockQuerier{})

	_, err := svc.RevertReviewer(context.Background(), "invalid-id")
	assert.ErrorIs(t, err, service.ErrInvalidCandidateID)
}

func TestRevertReviewer_NoCurrentReviewer_ReturnsErrNoReviewerToRevert(t *testing.T) {
	candidateID := mustScanUUID(t, "00000000-0000-0000-0000-000000000001")
	mockRepo := &mocks.MockQuerier{
		GetCurrentCandidateReviewerFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) (repository.CandidateReviewer, error) {
			assert.Equal(t, candidateID, inputCandidateID)
			return repository.CandidateReviewer{}, pgx.ErrNoRows
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, err := svc.RevertReviewer(context.Background(), "00000000-0000-0000-0000-000000000001")
	assert.ErrorIs(t, err, service.ErrNoReviewerToRevert)
}

func TestRevertReviewer_GetCurrentReviewerDBError_PropagatesError(t *testing.T) {
	expectedErr := errors.New("database unavailable")
	mockRepo := &mocks.MockQuerier{
		GetCurrentCandidateReviewerFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) (repository.CandidateReviewer, error) {
			return repository.CandidateReviewer{}, expectedErr
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, err := svc.RevertReviewer(context.Background(), "00000000-0000-0000-0000-000000000001")
	assert.ErrorIs(t, err, expectedErr)
}

func TestRevertReviewer_ReviewedOrNonPending_ReturnsErrReviewAlreadySubmitted(t *testing.T) {
	mockRepo := &mocks.MockQuerier{
		GetCurrentCandidateReviewerFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) (repository.CandidateReviewer, error) {
			return repository.CandidateReviewer{
				ReviewStatus: "suitable",
			}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, err := svc.RevertReviewer(context.Background(), "00000000-0000-0000-0000-000000000001")
	assert.ErrorIs(t, err, service.ErrReviewAlreadySubmitted)
}

func TestRevertReviewer_RemoveFails_ReturnsError(t *testing.T) {
	expectedErr := errors.New("remove failed")
	mockRepo := &mocks.MockQuerier{
		GetCurrentCandidateReviewerFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) (repository.CandidateReviewer, error) {
			return repository.CandidateReviewer{
				ReviewStatus: "pending",
			}, nil
		},
		RemoveCandidateReviewerFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) (int64, error) {
			return 0, expectedErr
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, err := svc.RevertReviewer(context.Background(), "00000000-0000-0000-0000-000000000001")
	assert.ErrorIs(t, err, expectedErr)
}

func TestRevertReviewer_RemoveAffectsNoRows_ReturnsErrReviewAlreadySubmitted(t *testing.T) {
	mockRepo := &mocks.MockQuerier{
		GetCurrentCandidateReviewerFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) (repository.CandidateReviewer, error) {
			return repository.CandidateReviewer{
				ReviewStatus: "pending",
			}, nil
		},
		RemoveCandidateReviewerFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) (int64, error) {
			return 0, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, err := svc.RevertReviewer(context.Background(), "00000000-0000-0000-0000-000000000001")
	assert.ErrorIs(t, err, service.ErrReviewAlreadySubmitted)
}

func TestRevertReviewer_ClearFails_ReturnsError(t *testing.T) {
	expectedErr := errors.New("clear failed")
	mockRepo := &mocks.MockQuerier{
		GetCurrentCandidateReviewerFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) (repository.CandidateReviewer, error) {
			return repository.CandidateReviewer{
				ReviewStatus: "pending",
			}, nil
		},
		RemoveCandidateReviewerFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) (int64, error) {
			return 1, nil
		},
		ClearCandidateReviewerFunc: func(ctx context.Context, id pgtype.UUID) error {
			return expectedErr
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, err := svc.RevertReviewer(context.Background(), "00000000-0000-0000-0000-000000000001")
	assert.ErrorIs(t, err, expectedErr)
}

func TestRevertReviewer_Success_CallsRemoveThenClearAndReturnsCandidate(t *testing.T) {
	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	candidateID := mustScanUUID(t, candidateIDStr)
	callOrder := make([]string, 0, 4)

	mockRepo := &mocks.MockQuerier{
		GetCurrentCandidateReviewerFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) (repository.CandidateReviewer, error) {
			callOrder = append(callOrder, "get_current")
			return repository.CandidateReviewer{
				ReviewStatus: "pending",
			}, nil
		},
		RemoveCandidateReviewerFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) (int64, error) {
			callOrder = append(callOrder, "remove")
			assert.Equal(t, candidateID, inputCandidateID)
			return 1, nil
		},
		ClearCandidateReviewerFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) error {
			callOrder = append(callOrder, "clear")
			assert.Equal(t, candidateID, inputCandidateID)
			return nil
		},
		DeleteNotificationsBySubjectIDAndEventTypeFunc: func(ctx context.Context, arg repository.DeleteNotificationsBySubjectIDAndEventTypeParams) error {
			callOrder = append(callOrder, "delete_notification")
			assert.Equal(t, candidateID, arg.SubjectID)
			assert.Equal(t, model.NotificationEventCandidateReviewerAssigned, arg.EventType)
			return nil
		},
		GetCandidateFunc: func(ctx context.Context, inputCandidateID pgtype.UUID) (repository.GetCandidateRow, error) {
			assert.Equal(t, candidateID, inputCandidateID)
			return repository.GetCandidateRow{
				ID:              candidateID,
				Name:            "John Doe",
				AppliedJobTitle: "Backend Engineer",
			}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	candidate, err := svc.RevertReviewer(context.Background(), candidateIDStr)
	assert.NoError(t, err)
	assert.Equal(t, candidateIDStr, candidate.ID)
	assert.Equal(t, "John Doe", candidate.Name)
	assert.Equal(t, []string{"get_current", "remove", "clear", "delete_notification"}, callOrder)
}

func mustScanUUID(t *testing.T, raw string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		t.Fatalf("failed to scan uuid %s: %v", raw, err)
	}
	return id
}

func TestUpdateCandidateResume_Success(t *testing.T) {
	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	candidateID := mustScanUUID(t, candidateIDStr)
	oldResumeURL := "/static/resumes/old-file.pdf"
	newResumeURL := "/static/resumes/new-file.pdf"

	updateCalled := false
	getCandidateCalls := 0
	mockRepo := &mocks.MockQuerier{
		UpdateCandidateResumeFunc: func(ctx context.Context, arg repository.UpdateCandidateResumeParams) (repository.Candidate, error) {
			updateCalled = true
			assert.Equal(t, candidateID, arg.ID)
			assert.Equal(t, newResumeURL, arg.ResumeUrl)
			return repository.Candidate{ID: candidateID}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			getCandidateCalls++
			resumeURL := oldResumeURL
			if getCandidateCalls > 1 {
				resumeURL = newResumeURL
			}
			return repository.GetCandidateRow{
				ID:              candidateID,
				Name:            "John Doe",
				ResumeUrl:       resumeURL,
				AppliedJobTitle: "Software Engineer",
			}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	candidate, gotOldResumeURL, err := svc.UpdateCandidateResume(context.Background(), candidateIDStr, newResumeURL)
	assert.NoError(t, err)
	assert.True(t, updateCalled)
	assert.Equal(t, candidateIDStr, candidate.ID)
	assert.Equal(t, newResumeURL, candidate.ResumeURL)
	assert.Equal(t, oldResumeURL, gotOldResumeURL)
}

func TestUpdateCandidateResume_InvalidID_ReturnsError(t *testing.T) {
	svc := service.NewCandidateService(&mocks.MockQuerier{})
	_, _, err := svc.UpdateCandidateResume(context.Background(), "invalid-id", "/static/resumes/test.pdf")
	assert.ErrorIs(t, err, service.ErrInvalidCandidateID)
}

func TestUpdateCandidateResume_CandidateNotFound_ReturnsError(t *testing.T) {
	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	candidateID := mustScanUUID(t, candidateIDStr)

	mockRepo := &mocks.MockQuerier{
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			assert.Equal(t, candidateID, id)
			return repository.GetCandidateRow{}, pgx.ErrNoRows
		},
		UpdateCandidateResumeFunc: func(ctx context.Context, arg repository.UpdateCandidateResumeParams) (repository.Candidate, error) {
			assert.Equal(t, candidateID, arg.ID)
			return repository.Candidate{}, pgx.ErrNoRows
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, _, err := svc.UpdateCandidateResume(context.Background(), candidateIDStr, "/static/resumes/test.pdf")
	assert.ErrorIs(t, err, service.ErrCandidateNotFound)
}

func TestUpdateCandidateResume_UpdateNoRows_ReturnsError(t *testing.T) {
	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	candidateID := mustScanUUID(t, candidateIDStr)

	mockRepo := &mocks.MockQuerier{
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{
				ID:        candidateID,
				Name:      "John Doe",
				ResumeUrl: "/static/resumes/old-file.pdf",
			}, nil
		},
		UpdateCandidateResumeFunc: func(ctx context.Context, arg repository.UpdateCandidateResumeParams) (repository.Candidate, error) {
			return repository.Candidate{}, pgx.ErrNoRows
		},
	}

	svc := service.NewCandidateService(mockRepo)
	_, _, err := svc.UpdateCandidateResume(context.Background(), candidateIDStr, "/static/resumes/new-file.pdf")
	assert.ErrorIs(t, err, service.ErrCandidateNotFound)
}

func TestUpdateCandidateResume_GetCandidateNoRowsFallbackAfterUpdate(t *testing.T) {
	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	candidateID := mustScanUUID(t, candidateIDStr)
	newResumeURL := "/static/resumes/new-file.pdf"
	getCandidateCalls := 0

	mockRepo := &mocks.MockQuerier{
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			getCandidateCalls++
			return repository.GetCandidateRow{}, pgx.ErrNoRows
		},
		UpdateCandidateResumeFunc: func(ctx context.Context, arg repository.UpdateCandidateResumeParams) (repository.Candidate, error) {
			return repository.Candidate{
				ID:              candidateID,
				Name:            "John Doe",
				Email:           "john@example.com",
				Phone:           "1234567890",
				ExperienceYears: 3,
				Education:       "BS",
				AppliedJobID:    mustScanUUID(t, "02020202-0202-0202-0202-020202020202"),
				Channel:         "LinkedIn",
				ResumeUrl:       newResumeURL,
				Status:          "new",
				AppliedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				ReviewStatus:    pgtype.Text{String: "pending", Valid: true},
			}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	candidate, oldResumeURL, err := svc.UpdateCandidateResume(context.Background(), candidateIDStr, newResumeURL)
	assert.NoError(t, err)
	assert.Equal(t, "", oldResumeURL)
	assert.Equal(t, candidateIDStr, candidate.ID)
	assert.Equal(t, newResumeURL, candidate.ResumeURL)
	assert.Equal(t, 2, getCandidateCalls)
}
