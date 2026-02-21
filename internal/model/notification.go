package model

import "time"

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

type NotificationUnreadCount struct {
	Count int64 `json:"count"`
}
