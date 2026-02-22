package notification

import (
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
