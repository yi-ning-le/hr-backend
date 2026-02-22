package model

import "time"

// Notification event types.
const (
	NotificationEventCandidateReviewerAssigned = "candidate_reviewer_assigned"
	NotificationEventInterviewAssigned         = "interview_assigned"
)

const (
	NotificationSubjectTypeCandidate = "candidate"
	NotificationSubjectTypeInterview = "interview"
)

type Notification struct {
	ID        string              `json:"id"`
	UserID    string              `json:"userId"`
	EventType string              `json:"eventType"`
	Subject   NotificationSubject `json:"subject"`
	Context   map[string]any      `json:"context,omitempty"`
	Content   NotificationContent `json:"content"`
	Action    *NotificationAction `json:"action,omitempty"`
	IsRead    bool                `json:"isRead"`
	CreatedAt time.Time           `json:"createdAt"`
}

type NotificationSubject struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type NotificationContent struct {
	TitleKey   string         `json:"titleKey"`
	MessageKey string         `json:"messageKey"`
	Params     map[string]any `json:"params,omitempty"`
}

type NotificationAction struct {
	Kind   string         `json:"kind"`
	Params map[string]any `json:"params,omitempty"`
}
