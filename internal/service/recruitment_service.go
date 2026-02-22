package service

import (
	"context"
	"log"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/utils"
)

type RecruitmentService struct {
	repo                  repository.Querier
	notificationPublisher *NotificationPublisher
}

func NewRecruitmentService(repo repository.Querier) *RecruitmentService {
	return &RecruitmentService{
		repo:                  repo,
		notificationPublisher: NewNotificationPublisher(repo),
	}
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
