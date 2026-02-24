package service

import (
	"context"
	"encoding/json"
	"errors"

	"hr-backend/internal/model"
	"hr-backend/internal/notification"
	"hr-backend/internal/repository"
	"hr-backend/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type NotificationPublisher struct {
	repo repository.Querier
}

func NewNotificationPublisher(repo repository.Querier) *NotificationPublisher {
	return &NotificationPublisher{repo: repo}
}

func (p *NotificationPublisher) PublishCandidateReviewerAssigned(
	ctx context.Context,
	reviewerEmployeeID pgtype.UUID,
	candidateID pgtype.UUID,
) error {
	return p.publishToEmployee(
		ctx,
		reviewerEmployeeID,
		model.NotificationEventCandidateReviewerAssigned,
		model.NotificationSubjectTypeCandidate,
		candidateID,
		notification.CandidateReviewerAssignedPayload{
			CandidateID: utils.UUIDToString(candidateID),
		},
	)
}

func (p *NotificationPublisher) PublishInterviewAssigned(
	ctx context.Context,
	interviewerEmployeeID pgtype.UUID,
	interviewID pgtype.UUID,
	candidateID pgtype.UUID,
) error {
	return p.publishToEmployee(
		ctx,
		interviewerEmployeeID,
		model.NotificationEventInterviewAssigned,
		model.NotificationSubjectTypeInterview,
		interviewID,
		notification.InterviewAssignedPayload{
			InterviewID: utils.UUIDToString(interviewID),
			CandidateID: utils.UUIDToString(candidateID),
		},
	)
}

func (p *NotificationPublisher) PublishReviewCompleted(
	ctx context.Context,
	recruiterUserID pgtype.UUID,
	candidateID pgtype.UUID,
	candidateName string,
	reviewStatus string,
	reviewerName string,
) error {
	if err := notification.ValidatePayload(
		model.NotificationEventReviewCompleted,
		model.NotificationSubjectTypeCandidate,
		notification.ReviewCompletedPayload{
			CandidateID:   utils.UUIDToString(candidateID),
			CandidateName: candidateName,
			ReviewStatus:  reviewStatus,
			ReviewerName:  reviewerName,
		},
	); err != nil {
		return err
	}

	payload := notification.ReviewCompletedPayload{
		CandidateID:   utils.UUIDToString(candidateID),
		CandidateName: candidateName,
		ReviewStatus:  reviewStatus,
		ReviewerName:  reviewerName,
	}

	contextJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = p.repo.CreateNotification(ctx, repository.CreateNotificationParams{
		UserID:      recruiterUserID,
		EventType:   model.NotificationEventReviewCompleted,
		SubjectType: model.NotificationSubjectTypeCandidate,
		SubjectID:   candidateID,
		Context:     contextJSON,
	})
	return err
}

func (p *NotificationPublisher) publishToEmployee(
	ctx context.Context,
	employeeID pgtype.UUID,
	eventType string,
	subjectType string,
	subjectID pgtype.UUID,
	payload any,
) error {
	employee, err := p.repo.GetEmployee(ctx, employeeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		return err
	}
	if !employee.UserID.Valid {
		return nil
	}

	if err := notification.ValidatePayload(eventType, subjectType, payload); err != nil {
		return err
	}

	contextJSON, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	_, err = p.repo.CreateNotification(ctx, repository.CreateNotificationParams{
		UserID:      employee.UserID,
		EventType:   eventType,
		SubjectType: subjectType,
		SubjectID:   subjectID,
		Context:     contextJSON,
	})
	return err
}
