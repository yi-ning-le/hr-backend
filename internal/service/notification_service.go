package service

import (
	"context"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/utils"
)

type NotificationService struct {
	repo repository.Querier
}

func NewNotificationService(repo repository.Querier) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) GetUserNotifications(ctx context.Context, userIDStr string, limit, offset int32) ([]model.Notification, error) {
	userID, err := utils.StringToUUID(userIDStr)
	if err != nil {
		return nil, err
	}

	params := repository.GetNotificationsByUserIdParams{
		UserID: userID,
		Limit:  limit,
		Offset: offset,
	}

	notifications, err := s.repo.GetNotificationsByUserId(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]model.Notification, len(notifications))
	for i, n := range notifications {
		result[i] = mapNotificationToModel(n)
	}
	return result, nil
}

func (s *NotificationService) GetUnreadCount(ctx context.Context, userIDStr string) (int64, error) {
	userID, err := utils.StringToUUID(userIDStr)
	if err != nil {
		return 0, err
	}

	return s.repo.GetUnreadNotificationCount(ctx, userID)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, idStr, userIDStr string) error {
	id, err := utils.StringToUUID(idStr)
	if err != nil {
		return err
	}

	userID, err := utils.StringToUUID(userIDStr)
	if err != nil {
		return err
	}

	params := repository.MarkNotificationAsReadParams{
		ID:     id,
		UserID: userID,
	}

	return s.repo.MarkNotificationAsRead(ctx, params)
}

func (s *NotificationService) MarkAllAsRead(ctx context.Context, userIDStr string) error {
	userID, err := utils.StringToUUID(userIDStr)
	if err != nil {
		return err
	}

	return s.repo.MarkAllNotificationsAsRead(ctx, userID)
}

func mapNotificationToModel(n repository.Notification) model.Notification {
	return model.Notification{
		ID:        utils.UUIDToString(n.ID),
		UserID:    utils.UUIDToString(n.UserID),
		Title:     n.Title,
		Message:   n.Message,
		Type:      n.Type,
		LinkUrl:   n.LinkUrl.String,
		IsRead:    n.IsRead,
		CreatedAt: n.CreatedAt.Time,
	}
}
