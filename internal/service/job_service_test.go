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

func TestCreateJob(t *testing.T) {
	mockRepo := &mocks.MockQuerier{
		CreateJobFunc: func(ctx context.Context, arg repository.CreateJobParams) (repository.Job, error) {
			return repository.Job{
				ID:             pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				Title:          arg.Title,
				Department:     arg.Department,
				HeadCount:      arg.HeadCount,
				OpenDate:       arg.OpenDate,
				JobDescription: arg.JobDescription,
				Status:         arg.Status,
				Note:           arg.Note,
			}, nil
		},
	}

	svc := service.NewJobService(mockRepo)

	input := model.JobInput{
		Title:          "Software Engineer",
		Department:     "Engineering",
		HeadCount:      2,
		OpenDate:       time.Now(),
		JobDescription: "Write Go code",
		Note:           "Urgent",
	}

	job, err := svc.CreateJob(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if job.Title != input.Title {
		t.Errorf("expected title %s, got %s", input.Title, job.Title)
	}
	if job.Status != "OPEN" {
		t.Errorf("expected default status OPEN, got %s", job.Status)
	}
}

func TestToggleJobStatus(t *testing.T) {
	jobIDStr := "01010101-0101-0101-0101-010101010101"
	var jobIDUUID pgtype.UUID
	jobIDUUID.Scan(jobIDStr)

	mockRepo := &mocks.MockQuerier{
		GetJobFunc: func(ctx context.Context, id pgtype.UUID) (repository.Job, error) {
			return repository.Job{
				ID:     jobIDUUID,
				Status: "OPEN",
			}, nil
		},
		UpdateJobStatusFunc: func(ctx context.Context, arg repository.UpdateJobStatusParams) (repository.Job, error) {
			return repository.Job{
				ID:     arg.ID,
				Status: arg.Status,
			}, nil
		},
	}

	svc := service.NewJobService(mockRepo)

	job, err := svc.ToggleJobStatus(context.Background(), jobIDStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if job.Status != "CLOSED" {
		t.Errorf("expected status CLOSED, got %s", job.Status)
	}
}
