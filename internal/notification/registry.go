package notification

import (
	"encoding/json"
	"errors"
	"fmt"

	"hr-backend/internal/model"
)

var (
	ErrUnknownEventType   = errors.New("unknown notification event type")
	ErrInvalidSubjectType = errors.New("invalid notification subject type")
	ErrInvalidContext     = errors.New("invalid notification context")
)

type CandidateReviewerAssignedPayload struct {
	CandidateID string `json:"candidateId"`
}

type InterviewAssignedPayload struct {
	InterviewID string `json:"interviewId"`
	CandidateID string `json:"candidateId"`
}

type ReviewCompletedPayload struct {
	CandidateID   string `json:"candidateId"`
	CandidateName string `json:"candidateName"`
	ReviewStatus  string `json:"reviewStatus"`
	ReviewerName  string `json:"reviewerName"`
}

func ValidatePayload(eventType, subjectType string, payload any) error {
	switch eventType {
	case model.NotificationEventCandidateReviewerAssigned:
		if subjectType != model.NotificationSubjectTypeCandidate {
			return fmt.Errorf("%w: expected %s got %s", ErrInvalidSubjectType, model.NotificationSubjectTypeCandidate, subjectType)
		}
		p, ok := payload.(CandidateReviewerAssignedPayload)
		if !ok {
			return fmt.Errorf("%w: candidate reviewer payload type mismatch", ErrInvalidContext)
		}
		if p.CandidateID == "" {
			return fmt.Errorf("%w: candidateId is required", ErrInvalidContext)
		}
		return nil
	case model.NotificationEventInterviewAssigned:
		if subjectType != model.NotificationSubjectTypeInterview {
			return fmt.Errorf("%w: expected %s got %s", ErrInvalidSubjectType, model.NotificationSubjectTypeInterview, subjectType)
		}
		p, ok := payload.(InterviewAssignedPayload)
		if !ok {
			return fmt.Errorf("%w: interview assigned payload type mismatch", ErrInvalidContext)
		}
		if p.InterviewID == "" {
			return fmt.Errorf("%w: interviewId is required", ErrInvalidContext)
		}
		if p.CandidateID == "" {
			return fmt.Errorf("%w: candidateId is required", ErrInvalidContext)
		}
		return nil
	case model.NotificationEventReviewCompleted:
		if subjectType != model.NotificationSubjectTypeCandidate {
			return fmt.Errorf("%w: expected %s got %s", ErrInvalidSubjectType, model.NotificationSubjectTypeCandidate, subjectType)
		}
		p, ok := payload.(ReviewCompletedPayload)
		if !ok {
			return fmt.Errorf("%w: review completed payload type mismatch", ErrInvalidContext)
		}
		if p.CandidateID == "" {
			return fmt.Errorf("%w: candidateId is required", ErrInvalidContext)
		}
		return nil
	default:
		return fmt.Errorf("%w: %s", ErrUnknownEventType, eventType)
	}
}

func BuildPresentation(eventType, subjectID string, context json.RawMessage) (model.NotificationContent, *model.NotificationAction) {
	// Common logic to unmarshal context into map for simple template substitution
	var contextMap map[string]any
	if len(context) > 0 {
		_ = json.Unmarshal(context, &contextMap)
	}

	switch eventType {
	case model.NotificationEventCandidateReviewerAssigned:
		content := model.NotificationContent{
			TitleKey:   "notifications.events.candidate_reviewer_assigned.title",
			MessageKey: "notifications.events.candidate_reviewer_assigned.message",
			Params:     contextMap,
		}
		candidateID := subjectID
		var payload CandidateReviewerAssignedPayload
		if err := json.Unmarshal(context, &payload); err == nil && payload.CandidateID != "" {
			candidateID = payload.CandidateID
		}
		return content, &model.NotificationAction{
			Kind:   "candidateReview",
			Params: map[string]any{"candidateId": candidateID},
		}
	case model.NotificationEventInterviewAssigned:
		content := model.NotificationContent{
			TitleKey:   "notifications.events.interview_assigned.title",
			MessageKey: "notifications.events.interview_assigned.message",
			Params:     contextMap,
		}
		interviewID := subjectID
		var payload InterviewAssignedPayload
		if err := json.Unmarshal(context, &payload); err == nil && payload.InterviewID != "" {
			interviewID = payload.InterviewID
		}
		return content, &model.NotificationAction{
			Kind:   "interviewDetail",
			Params: map[string]any{"interviewId": interviewID},
		}
	case model.NotificationEventReviewCompleted:
		content := model.NotificationContent{
			TitleKey:   "notifications.events.review_completed.title",
			MessageKey: "notifications.events.review_completed.message",
			Params:     contextMap,
		}
		candidateID := subjectID
		var payload ReviewCompletedPayload
		if err := json.Unmarshal(context, &payload); err == nil && payload.CandidateID != "" {
			candidateID = payload.CandidateID
		}
		return content, &model.NotificationAction{
			Kind:   "reviewFinished",
			Params: map[string]any{"candidateId": candidateID},
		}
	default:
		return model.NotificationContent{
			TitleKey:   "notifications.events.generic.title",
			MessageKey: "notifications.events.generic.message",
		}, nil
	}
}
