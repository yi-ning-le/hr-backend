package service

import (
	"context"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type JobService struct {
	repo repository.Querier
}

func NewJobService(repo repository.Querier) *JobService {
	return &JobService{repo: repo}
}

func (s *JobService) CreateJob(ctx context.Context, input model.JobInput) (*model.JobPosition, error) {
	status := input.Status
	if status == "" {
		status = "OPEN"
	}

	params := repository.CreateJobParams{
		Title:          input.Title,
		Department:     input.Department,
		HeadCount:      int32(input.HeadCount),
		OpenDate:       pgtype.Timestamptz{Time: input.OpenDate, Valid: true},
		JobDescription: input.JobDescription,
		Note:           pgtype.Text{String: input.Note, Valid: input.Note != ""},
		Status:         status,
	}

	job, err := s.repo.CreateJob(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapJobToModel(job), nil
}

func (s *JobService) ListJobs(ctx context.Context) ([]model.JobPosition, error) {
	jobs, err := s.repo.ListJobs(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]model.JobPosition, len(jobs))
	for i, j := range jobs {
		result[i] = *mapJobToModel(j)
	}
	return result, nil
}

func (s *JobService) GetJob(ctx context.Context, id string) (*model.JobPosition, error) {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	job, err := s.repo.GetJob(ctx, uuid)
	if err != nil {
		return nil, err
	}

	return mapJobToModel(job), nil
}

func (s *JobService) UpdateJob(ctx context.Context, id string, input model.JobInput) (*model.JobPosition, error) {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	params := repository.UpdateJobParams{
		ID:             uuid,
		Title:          input.Title,
		Department:     input.Department,
		HeadCount:      int32(input.HeadCount),
		OpenDate:       pgtype.Timestamptz{Time: input.OpenDate, Valid: true},
		JobDescription: input.JobDescription,
		Note:           pgtype.Text{String: input.Note, Valid: input.Note != ""},
		Status:         input.Status,
	}

	job, err := s.repo.UpdateJob(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapJobToModel(job), nil
}

func (s *JobService) ToggleJobStatus(ctx context.Context, id string) (*model.JobPosition, error) {
	// First get current status to toggle
	// Or simpler: The API spec says "Toggle" but the endpoint is PATCH /jobs/{id}/status which implies sending a new status?
	// The OpenAPI spec description says "Toggle job status (OPEN/CLOSED)", but the schema doesn't define a body for PATCH?
	// Wait, looking at OpenAPI:
	// /jobs/{id}/status PATCH "Toggle job status"
	// Response: 200 JobPosition.
	// It does NOT have a requestBody. So it really is a toggle.

	currentJob, err := s.GetJob(ctx, id)
	if err != nil {
		return nil, err
	}

	newStatus := "CLOSED"
	if currentJob.Status == "CLOSED" {
		newStatus = "OPEN"
	}

	uuid, _ := utils.StringToUUID(id) // Already validated in GetJob

	params := repository.UpdateJobStatusParams{
		ID:     uuid,
		Status: newStatus,
	}

	updatedJob, err := s.repo.UpdateJobStatus(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapJobToModel(updatedJob), nil
}

func (s *JobService) DeleteJob(ctx context.Context, id string) error {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return err
	}
	return s.repo.DeleteJob(ctx, uuid)
}

// Helper to convert DB model to API model
func mapJobToModel(j repository.Job) *model.JobPosition {
	return &model.JobPosition{
		ID:             utils.UUIDToString(j.ID),
		Title:          j.Title,
		Department:     j.Department,
		HeadCount:      int(j.HeadCount),
		OpenDate:       j.OpenDate.Time,
		JobDescription: j.JobDescription,
		Note:           j.Note.String,
		Status:         j.Status,
	}
}
