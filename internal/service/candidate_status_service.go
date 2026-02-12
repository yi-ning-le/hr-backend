package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"hr-backend/internal/repository"
	"hr-backend/internal/utils"
)

var ErrInvalidStatusID = errors.New("invalid status id")

type CandidateStatusService struct {
	repo repository.Querier
}

func NewCandidateStatusService(repo repository.Querier) *CandidateStatusService {
	return &CandidateStatusService{repo: repo}
}

func (s *CandidateStatusService) ListStatuses(ctx context.Context) ([]repository.CandidateStatus, error) {
	return s.repo.ListCandidateStatuses(ctx)
}

func (s *CandidateStatusService) CreateStatus(ctx context.Context, name, color string) (*repository.CandidateStatus, error) {
	// Generate slug from name
	slug := strings.ToLower(strings.ReplaceAll(name, " ", "-"))

	// Get max order (simplified: fetch all and add 1, or just let DB default if it was serial, but here we scan all)
	statuses, err := s.repo.ListCandidateStatuses(ctx)
	if err != nil {
		return nil, err
	}
	nextOrder := int32(len(statuses) + 1)

	status, err := s.repo.CreateCandidateStatus(ctx, repository.CreateCandidateStatusParams{
		Name:      name,
		Slug:      slug,
		Type:      "custom",
		SortOrder: nextOrder,
		Color:     color,
	})
	if err != nil {
		return nil, err
	}
	return &status, nil
}

func (s *CandidateStatusService) UpdateStatus(ctx context.Context, id string, name, color string) (*repository.CandidateStatus, error) {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	status, err := s.repo.GetCandidateStatus(ctx, uuid)
	if err != nil {
		return nil, err
	}

	// Prevent renaming system statuses? Maybe allow it but keep the slug same.
	// Current implementation allows renaming.

	updated, err := s.repo.UpdateCandidateStatusFields(ctx, repository.UpdateCandidateStatusFieldsParams{
		ID:    status.ID,
		Name:  name,
		Color: color,
	})
	if err != nil {
		return nil, err
	}
	return &updated, nil
}

func (s *CandidateStatusService) DeleteStatus(ctx context.Context, id string) error {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return err
	}

	status, err := s.repo.GetCandidateStatus(ctx, uuid)
	if err != nil {
		return err
	}

	if status.Type == "system" {
		return errors.New("cannot delete system status")
	}

	// TODO: Check if any candidates are using this status before deleting

	return s.repo.DeleteCandidateStatus(ctx, uuid)
}

func (s *CandidateStatusService) ReorderStatuses(ctx context.Context, ids []string) error {
	parsedIDs := make([]string, len(ids))
	for i, id := range ids {
		if _, err := utils.StringToUUID(id); err != nil {
			return fmt.Errorf("%w at index %d", ErrInvalidStatusID, i)
		}
		parsedIDs[i] = id
	}

	for i, id := range parsedIDs {
		uuid, _ := utils.StringToUUID(id)
		err := s.repo.UpdateCandidateStatusOrder(ctx, repository.UpdateCandidateStatusOrderParams{
			ID:        uuid,
			SortOrder: int32(i + 1),
		})
		if err != nil {
			return err
		}
	}
	return nil
}
