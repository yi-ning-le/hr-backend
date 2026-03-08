package service

import (
	"context"
	"errors"
	"log"
	"strings"

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
	ErrInterviewResultRequired       = errors.New("result is required when completing an interview")
	ErrInvalidInterviewResult        = errors.New("result must be PASS or FAIL when completing an interview")
	ErrCancelledInterviewPayload     = errors.New("result/comment are not allowed when cancelling an interview")
	ErrInvalidInterviewStatusPayload = errors.New("invalid interview status payload")
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

func (s *RecruitmentService) CompleteInterview(
	ctx context.Context,
	interviewID pgtype.UUID,
	interview repository.Interview,
	result string,
	comment string,
) (repository.Interview, error) {
	normalizedResult, err := normalizeInterviewResult(result)
	if err != nil {
		return repository.Interview{}, err
	}

	updated, err := s.repo.UpdateInterviewStatus(ctx, repository.UpdateInterviewStatusParams{
		ID:     interviewID,
		Status: "COMPLETED",
	})
	if err != nil {
		return repository.Interview{}, err
	}

	// Create a candidate comment with interview result type
	commentType := "interview_pass"
	if normalizedResult == "FAIL" {
		commentType = "interview_fail"
	}
	commentContent := strings.TrimSpace(comment)
	if commentContent == "" {
		commentContent = normalizedResult
	}
	if _, createErr := s.repo.CreateCandidateComment(ctx, repository.CreateCandidateCommentParams{
		CandidateID: interview.CandidateID,
		AuthorID:    interview.InterviewerID,
		Content:     commentContent,
		CommentType: commentType,
	}); createErr != nil {
		log.Printf(
			"warn: failed to create interview result comment interview_id=%s candidate_id=%s err=%v",
			utils.UUIDToString(interviewID),
			utils.UUIDToString(interview.CandidateID),
			createErr,
		)
	}

	// Delete the interview_assigned notification for the interviewer
	if deleteErr := s.repo.DeleteNotificationsBySubjectIDAndEventType(ctx, repository.DeleteNotificationsBySubjectIDAndEventTypeParams{
		SubjectID: interviewID,
		EventType: model.NotificationEventInterviewAssigned,
	}); deleteErr != nil {
		log.Printf(
			"warn: failed to delete interview_assigned notification interview_id=%s err=%v",
			utils.UUIDToString(interviewID),
			deleteErr,
		)
	}

	// Notify the recruiter who created the interview
	if interview.CreatedByUserID.Valid {
		candidateName := ""
		interviewerName := ""

		candidate, candidateErr := s.repo.GetCandidate(ctx, interview.CandidateID)
		if candidateErr == nil {
			candidateName = candidate.Name
		}

		interviewer, interviewerErr := s.repo.GetEmployee(ctx, interview.InterviewerID)
		if interviewerErr == nil {
			interviewerName = strings.TrimSpace(interviewer.FirstName + " " + interviewer.LastName)
		}

		if notifyErr := s.notificationPublisher.PublishInterviewCompleted(
			ctx,
			interview.CreatedByUserID,
			interviewID,
			interview.CandidateID,
			candidateName,
			interviewerName,
			normalizedResult,
		); notifyErr != nil {
			log.Printf(
				"warn: failed to publish notification event=%s recruiter_user_id=%s interview_id=%s err=%v",
				model.NotificationEventInterviewCompleted,
				utils.UUIDToString(interview.CreatedByUserID),
				utils.UUIDToString(interviewID),
				notifyErr,
			)
		}
	}

	return updated, nil
}

func (s *RecruitmentService) UpdateInterviewStatus(
	ctx context.Context,
	interviewID pgtype.UUID,
	interview repository.Interview,
	input model.UpdateInterviewStatusInput,
) (repository.Interview, error) {
	switch input.Status {
	case "COMPLETED":
		if input.Result == nil {
			return repository.Interview{}, ErrInterviewResultRequired
		}
		comment := ""
		if input.Comment != nil {
			comment = *input.Comment
		}
		return s.CompleteInterview(ctx, interviewID, interview, *input.Result, comment)
	case "CANCELLED":
		if input.Result != nil || input.Comment != nil {
			return repository.Interview{}, ErrCancelledInterviewPayload
		}
		return s.repo.UpdateInterviewStatus(ctx, repository.UpdateInterviewStatusParams{
			ID:     interviewID,
			Status: "CANCELLED",
		})
	default:
		return repository.Interview{}, ErrInvalidInterviewStatusPayload
	}
}

func normalizeInterviewResult(raw string) (string, error) {
	result := strings.ToUpper(strings.TrimSpace(raw))
	if result != "PASS" && result != "FAIL" {
		return "", ErrInvalidInterviewResult
	}
	return result, nil
}
