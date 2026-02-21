package model

import "time"

// Notification Types
const (
	NotificationTypeInfo    = "info"
	NotificationTypeSuccess = "success"
	NotificationTypeWarning = "warning"
	NotificationTypeError   = "error"
	// Domain specific
	NotificationTypeSystem    = "system"
	NotificationTypeCandidate = "candidate" // For candidate updates
	NotificationTypeInterview = "interview" // For interview updates
)

type Notification struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	LinkUrl   string    `json:"linkUrl,omitempty"`
	IsRead    bool      `json:"isRead"`
	CreatedAt time.Time `json:"createdAt"`
}
