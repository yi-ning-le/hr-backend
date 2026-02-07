package service

import (
	"context"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type CandidateService struct {
	repo repository.Querier
}

func NewCandidateService(repo repository.Querier) *CandidateService {
	return &CandidateService{repo: repo}
}

func (s *CandidateService) CreateCandidate(ctx context.Context, input model.CandidateInput) (*model.Candidate, error) {
	jobUUID, err := utils.StringToUUID(input.AppliedJobID)
	if err != nil {
		return nil, err
	}

	status := input.Status
	if status == "" {
		status = "new"
	}

	params := repository.CreateCandidateParams{
		Name:            input.Name,
		Avatar:          pgtype.Text{String: input.Avatar, Valid: input.Avatar != ""},
		Email:           input.Email,
		Phone:           input.Phone,
		ExperienceYears: int32(input.ExperienceYears),
		Education:       input.Education,
		AppliedJobID:    jobUUID,
		Channel:         input.Channel,
		ResumeUrl:       input.ResumeURL,
		Status:          status,
		Note:            pgtype.Text{String: input.Note, Valid: input.Note != ""},
		AppliedAt:       pgtype.Timestamptz{Time: input.AppliedAt, Valid: true},
	}

	candidate, err := s.repo.CreateCandidate(ctx, params)
	if err != nil {
		return nil, err
	}

	// We need to fetch the job title to complete the return model,
	// but CreateCandidate returns the candidate row which doesn't have the title.
	// We can either fetch the job separately or just use the input job title (if we trust it or if it's optional).
	// The generated query `GetCandidate` joins with jobs. Let's use that to return the full object.
	return s.GetCandidate(ctx, utils.UUIDToString(candidate.ID))
}

func (s *CandidateService) ListCandidates(ctx context.Context, jobIDFilter string) ([]model.Candidate, error) {
	var filterUUID pgtype.UUID
	if jobIDFilter != "" {
		var err error
		filterUUID, err = utils.StringToUUID(jobIDFilter)
		if err != nil {
			return nil, err // Or ignore filter? API should probably validation error.
		}
	} else {
		filterUUID.Valid = false // Null UUID for "all"
	}

	rows, err := s.repo.ListCandidates(ctx, filterUUID)
	if err != nil {
		return nil, err
	}

	result := make([]model.Candidate, len(rows))
	for i, r := range rows {
		result[i] = mapCandidateRowToModel(r)
	}
	return result, nil
}

func (s *CandidateService) GetCandidate(ctx context.Context, id string) (*model.Candidate, error) {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	row, err := s.repo.GetCandidate(ctx, uuid)
	if err != nil {
		return nil, err
	}

	// GetCandidate returns a Row with joined fields
	return &model.Candidate{
		ID:              utils.UUIDToString(row.ID),
		Name:            row.Name,
		Avatar:          row.Avatar.String,
		Email:           row.Email,
		Phone:           row.Phone,
		ExperienceYears: int(row.ExperienceYears),
		Education:       row.Education,
		AppliedJobID:    utils.UUIDToString(row.AppliedJobID),
		AppliedJobTitle: row.AppliedJobTitle,
		Channel:         row.Channel,
		ResumeURL:       row.ResumeUrl,
		Status:          row.Status,
		Note:            row.Note.String,
		AppliedAt:       row.AppliedAt.Time,
	}, nil
}

func (s *CandidateService) UpdateCandidate(ctx context.Context, id string, input model.CandidateInput) (*model.Candidate, error) {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	jobUUID, err := utils.StringToUUID(input.AppliedJobID)
	if err != nil {
		return nil, err
	}

	params := repository.UpdateCandidateParams{
		ID:              uuid,
		Name:            input.Name,
		Avatar:          pgtype.Text{String: input.Avatar, Valid: input.Avatar != ""},
		Email:           input.Email,
		Phone:           input.Phone,
		ExperienceYears: int32(input.ExperienceYears),
		Education:       input.Education,
		AppliedJobID:    jobUUID,
		Channel:         input.Channel,
		ResumeUrl:       input.ResumeURL,
		Status:          input.Status,
		Note:            pgtype.Text{String: input.Note, Valid: input.Note != ""},
		AppliedAt:       pgtype.Timestamptz{Time: input.AppliedAt, Valid: true},
	}

	_, err = s.repo.UpdateCandidate(ctx, params)
	if err != nil {
		return nil, err
	}

	return s.GetCandidate(ctx, id)
}

func (s *CandidateService) UpdateStatus(ctx context.Context, id string, status string) (*model.Candidate, error) {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.UpdateCandidateStatus(ctx, repository.UpdateCandidateStatusParams{
		ID:     uuid,
		Status: status,
	})
	if err != nil {
		return nil, err
	}
	return s.GetCandidate(ctx, id)
}

func (s *CandidateService) UpdateNote(ctx context.Context, id string, note string) (*model.Candidate, error) {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.UpdateCandidateNote(ctx, repository.UpdateCandidateNoteParams{
		ID:   uuid,
		Note: pgtype.Text{String: note, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return s.GetCandidate(ctx, id)
}

func (s *CandidateService) UpdateResume(ctx context.Context, id string, resumeUrl string) (*model.Candidate, error) {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.UpdateCandidateResume(ctx, repository.UpdateCandidateResumeParams{
		ID:        uuid,
		ResumeUrl: resumeUrl,
	})
	if err != nil {
		return nil, err
	}
	return s.GetCandidate(ctx, id)
}

func (s *CandidateService) DeleteCandidate(ctx context.Context, id string) error {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return err
	}
	return s.repo.DeleteCandidate(ctx, uuid)
}

func mapCandidateRowToModel(row repository.ListCandidatesRow) model.Candidate {
	return model.Candidate{
		ID:              utils.UUIDToString(row.ID),
		Name:            row.Name,
		Avatar:          row.Avatar.String,
		Email:           row.Email,
		Phone:           row.Phone,
		ExperienceYears: int(row.ExperienceYears),
		Education:       row.Education,
		AppliedJobID:    utils.UUIDToString(row.AppliedJobID),
		AppliedJobTitle: row.AppliedJobTitle,
		Channel:         row.Channel,
		ResumeURL:       row.ResumeUrl,
		Status:          row.Status,
		Note:            row.Note.String,
		AppliedAt:       row.AppliedAt.Time,
	}
}
