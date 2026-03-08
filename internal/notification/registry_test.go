package notification

import (
	"encoding/json"
	"testing"

	"hr-backend/internal/model"
)

func TestValidate(t *testing.T) {
	err := ValidatePayload(
		model.NotificationEventCandidateReviewerAssigned,
		model.NotificationSubjectTypeCandidate,
		CandidateReviewerAssignedPayload{
			CandidateID: "11111111-1111-1111-1111-111111111111",
		},
	)
	if err != nil {
		t.Fatalf("expected valid payload, got error: %v", err)
	}
}

func TestValidatePayload_InvalidSubjectType(t *testing.T) {
	err := ValidatePayload(
		model.NotificationEventInterviewAssigned,
		model.NotificationSubjectTypeCandidate,
		InterviewAssignedPayload{
			InterviewID: "22222222-2222-2222-2222-222222222222",
			CandidateID: "11111111-1111-1111-1111-111111111111",
		},
	)
	if err == nil {
		t.Fatalf("expected subject type validation error")
	}
}

func TestBuildPresentation_UnknownEventFallback(t *testing.T) {
	content, action := BuildPresentation("future_event", "x", nil)
	if content.TitleKey != "notifications.events.generic.title" {
		t.Fatalf("unexpected title fallback: %s", content.TitleKey)
	}
	if action != nil {
		t.Fatalf("expected nil action for unknown event")
	}
}

func TestValidatePayload_ReviewCompleted(t *testing.T) {
	err := ValidatePayload(
		model.NotificationEventReviewCompleted,
		model.NotificationSubjectTypeCandidate,
		ReviewCompletedPayload{
			CandidateID:  "11111111-1111-1111-1111-111111111111",
			ReviewStatus: "suitable",
			ReviewerName: "Alice Lee",
		},
	)
	if err != nil {
		t.Fatalf("expected valid review_completed payload, got error: %v", err)
	}
}

func TestBuildPresentation_ReviewCompleted(t *testing.T) {
	content, action := BuildPresentation(
		model.NotificationEventReviewCompleted,
		"11111111-1111-1111-1111-111111111111",
		json.RawMessage(`{
			"candidateId":  "11111111-1111-1111-1111-111111111111",
			"reviewStatus": "unsuitable",
			"reviewerName": "Alice Lee"
		}`),
	)

	if content.TitleKey != "notifications.events.review_completed.title" {
		t.Fatalf("unexpected title key: %s", content.TitleKey)
	}
	if content.MessageKey != "notifications.events.review_completed.message" {
		t.Fatalf("unexpected message key: %s", content.MessageKey)
	}
	if action == nil {
		t.Fatalf("expected non-nil action")
	}
	if action.Kind != "reviewFinished" {
		t.Fatalf("unexpected action kind: %s", action.Kind)
	}
	if action.Params["candidateId"] != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("unexpected candidateId param: %v", action.Params["candidateId"])
	}
}

func TestValidatePayload_InterviewCompleted(t *testing.T) {
	err := ValidatePayload(
		model.NotificationEventInterviewCompleted,
		model.NotificationSubjectTypeInterview,
		InterviewCompletedPayload{
			InterviewID:     "22222222-2222-2222-2222-222222222222",
			CandidateID:     "11111111-1111-1111-1111-111111111111",
			CandidateName:   "John Doe",
			InterviewerName: "Alice Lee",
		},
	)
	if err != nil {
		t.Fatalf("expected valid interview_completed payload, got error: %v", err)
	}
}

func TestValidatePayload_InterviewCompleted_InvalidSubjectType(t *testing.T) {
	err := ValidatePayload(
		model.NotificationEventInterviewCompleted,
		model.NotificationSubjectTypeCandidate,
		InterviewCompletedPayload{
			InterviewID:     "22222222-2222-2222-2222-222222222222",
			CandidateID:     "11111111-1111-1111-1111-111111111111",
			CandidateName:   "John Doe",
			InterviewerName: "Alice Lee",
		},
	)
	if err == nil {
		t.Fatalf("expected subject type validation error")
	}
}

func TestValidatePayload_InterviewCompleted_MissingInterviewID(t *testing.T) {
	err := ValidatePayload(
		model.NotificationEventInterviewCompleted,
		model.NotificationSubjectTypeInterview,
		InterviewCompletedPayload{
			CandidateID:     "11111111-1111-1111-1111-111111111111",
			CandidateName:   "John Doe",
			InterviewerName: "Alice Lee",
		},
	)
	if err == nil {
		t.Fatalf("expected validation error for missing interviewId")
	}
}

func TestBuildPresentation_InterviewCompleted(t *testing.T) {
	content, action := BuildPresentation(
		model.NotificationEventInterviewCompleted,
		"22222222-2222-2222-2222-222222222222",
		json.RawMessage(`{
			"interviewId":     "22222222-2222-2222-2222-222222222222",
			"candidateId":     "11111111-1111-1111-1111-111111111111",
			"candidateName":   "John Doe",
			"interviewerName": "Alice Lee"
		}`),
	)

	if content.TitleKey != "notifications.events.interview_completed.title" {
		t.Fatalf("unexpected title key: %s", content.TitleKey)
	}
	if content.MessageKey != "notifications.events.interview_completed.message" {
		t.Fatalf("unexpected message key: %s", content.MessageKey)
	}
	if action == nil {
		t.Fatalf("expected non-nil action")
	}
	if action.Kind != "interviewCompleted" {
		t.Fatalf("unexpected action kind: %s", action.Kind)
	}
	if action.Params["candidateId"] != "11111111-1111-1111-1111-111111111111" {
		t.Fatalf("unexpected candidateId param: %v", action.Params["candidateId"])
	}
}
