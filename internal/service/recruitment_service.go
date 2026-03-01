package service

import (
	"context"
	"errors"
	"log"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type RecruitmentService struct {
	repo                  repository.Querier
	notificationPublisher *NotificationPublisher
}

var (
	ErrInterviewNotPendingOrNotFound = errors.New("interview not pending or not found")
)

func NewRecruitmentService(repo repository.Querier) *RecruitmentService {
	return &RecruitmentService{
		repo:                  repo,
		notificationPublisher: NewNotificationPublisher(repo),
	}
}

func (s *RecruitmentService) DeleteInterview(ctx context.Context, interviewID pgtype.UUID) error {
	deletedRows, err := s.repo.DeleteInterview(ctx, interviewID)
	if err != nil {
		return err
	}
	if deletedRows == 0 {
		return ErrInterviewNotPendingOrNotFound
	}

	if err := s.repo.DeleteNotificationsBySubjectIDAndEventType(ctx, repository.DeleteNotificationsBySubjectIDAndEventTypeParams{
		SubjectID: interviewID,
		EventType: model.NotificationEventInterviewAssigned,
	}); err != nil {
		log.Printf(
			"warn: failed to delete notifications for interview_id=%s err=%v",
			utils.UUIDToString(interviewID),
			err,
		)
	}

	return nil
}

func (s *RecruitmentService) CreateInterview(
	ctx context.Context,
	params repository.CreateInterviewParams,
) (repository.CreateInterviewRow, error) {
	interview, err := s.repo.CreateInterview(ctx, params)
	if err != nil {
		return repository.CreateInterviewRow{}, err
	}

	if err := s.notificationPublisher.PublishInterviewAssigned(
		ctx,
		params.InterviewerID,
		interview.ID,
		interview.CandidateID,
	); err != nil {
		log.Printf(
			"warn: failed to publish notification event=%s interviewer_id=%s interview_id=%s candidate_id=%s err=%v",
			model.NotificationEventInterviewAssigned,
			utils.UUIDToString(params.InterviewerID),
			utils.UUIDToString(interview.ID),
			utils.UUIDToString(interview.CandidateID),
			err,
		)
	}

	return interview, nil
}
