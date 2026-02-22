package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"hr-backend/internal/repository"
	"hr-backend/internal/service"
	"hr-backend/test/mocks"

	"github.com/jackc/pgx/v5/pgtype"
)

func TestRecruitmentService_CreateInterview_PublishesNotification(t *testing.T) {
	candidateID := parseUUID(t, "11111111-1111-1111-1111-111111111111")
	interviewerID := parseUUID(t, "22222222-2222-2222-2222-222222222222")
	jobID := parseUUID(t, "33333333-3333-3333-3333-333333333333")
	interviewID := parseUUID(t, "44444444-4444-4444-4444-444444444444")

	created := false
	notified := false
	mockRepo := &mocks.MockQuerier{
		CreateInterviewFunc: func(ctx context.Context, arg repository.CreateInterviewParams) (repository.CreateInterviewRow, error) {
			created = true
			return repository.CreateInterviewRow{
				ID:               interviewID,
				CandidateID:      candidateID,
				InterviewerID:    arg.InterviewerID,
				JobID:            arg.JobID,
				ScheduledTime:    arg.ScheduledTime,
				ScheduledEndTime: arg.ScheduledEndTime,
				Status:           arg.Status,
				CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
		GetEmployeeFunc: func(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{
				ID:     interviewerID,
				UserID: candidateID, // reuse UUID as mock user ID
			}, nil
		},
		CreateNotificationFunc: func(ctx context.Context, arg repository.CreateNotificationParams) (repository.Notification, error) {
			notified = true
			return repository.Notification{}, nil
		},
	}

	svc := service.NewRecruitmentService(mockRepo)
	_, err := svc.CreateInterview(context.Background(), repository.CreateInterviewParams{
		ID:               candidateID,
		InterviewerID:    interviewerID,
		JobID:            jobID,
		ScheduledTime:    pgtype.Timestamptz{Time: time.Now().Add(2 * time.Hour), Valid: true},
		ScheduledEndTime: pgtype.Timestamptz{Time: time.Now().Add(3 * time.Hour), Valid: true},
		Status:           "PENDING",
	})
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if !created {
		t.Fatalf("expected CreateInterview to be called")
	}
	if !notified {
		t.Fatalf("expected CreateNotification to be called")
	}
}

func TestRecruitmentService_CreateInterview_CreateError(t *testing.T) {
	boom := errors.New("boom")
	candidateID := parseUUID(t, "11111111-1111-1111-1111-111111111111")
	interviewerID := parseUUID(t, "22222222-2222-2222-2222-222222222222")
	jobID := parseUUID(t, "33333333-3333-3333-3333-333333333333")

	notificationCalled := false
	mockRepo := &mocks.MockQuerier{
		CreateInterviewFunc: func(ctx context.Context, arg repository.CreateInterviewParams) (repository.CreateInterviewRow, error) {
			return repository.CreateInterviewRow{}, boom
		},
		CreateNotificationFunc: func(ctx context.Context, arg repository.CreateNotificationParams) (repository.Notification, error) {
			notificationCalled = true
			return repository.Notification{}, nil
		},
	}

	svc := service.NewRecruitmentService(mockRepo)
	_, err := svc.CreateInterview(context.Background(), repository.CreateInterviewParams{
		ID:               candidateID,
		InterviewerID:    interviewerID,
		JobID:            jobID,
		ScheduledTime:    pgtype.Timestamptz{Time: time.Now().Add(2 * time.Hour), Valid: true},
		ScheduledEndTime: pgtype.Timestamptz{Time: time.Now().Add(3 * time.Hour), Valid: true},
		Status:           "PENDING",
	})
	if !errors.Is(err, boom) {
		t.Fatalf("expected error %v, got %v", boom, err)
	}
	if notificationCalled {
		t.Fatalf("notification should not be called when create interview fails")
	}
}

func parseUUID(t *testing.T, raw string) pgtype.UUID {
	t.Helper()
	var id pgtype.UUID
	if err := id.Scan(raw); err != nil {
		t.Fatalf("failed to parse uuid %s: %v", raw, err)
	}
	return id
}
