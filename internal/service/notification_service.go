package service

import (
	"context"
	"encoding/json"
	"errors"

	"hr-backend/internal/model"
	"hr-backend/internal/notification"
	"hr-backend/internal/repository"
	"hr-backend/internal/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

var ErrInvalidUUID = errors.New("invalid uuid")

type NotificationService struct {
	repo repository.Querier
}

func NewNotificationService(repo repository.Querier) *NotificationService {
	return &NotificationService{repo: repo}
}

func (s *NotificationService) GetUserNotifications(ctx context.Context, userIDStr string, limit, offset int32) ([]model.Notification, error) {
	userID, err := parseUUID(userIDStr)
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
	userID, err := parseUUID(userIDStr)
	if err != nil {
		return 0, err
	}

	return s.repo.GetUnreadNotificationCount(ctx, userID)
}

func (s *NotificationService) MarkAsRead(ctx context.Context, idStr, userIDStr string) error {
	id, err := parseUUID(idStr)
	if err != nil {
		return err
	}

	userID, err := parseUUID(userIDStr)
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
	userID, err := parseUUID(userIDStr)
	if err != nil {
		return err
	}

	return s.repo.MarkAllNotificationsAsRead(ctx, userID)
}

func (s *NotificationService) DeleteNotification(ctx context.Context, idStr, userIDStr string) error {
	id, err := parseUUID(idStr)
	if err != nil {
		return err
	}

	userID, err := parseUUID(userIDStr)
	if err != nil {
		return err
	}

	return s.repo.DeleteNotification(ctx, repository.DeleteNotificationParams{
		ID:     id,
		UserID: userID,
	})
}

func parseUUID(raw string) (id pgtype.UUID, err error) {
	parsed, parseErr := utils.StringToUUID(raw)
	if parseErr != nil {
		return pgtype.UUID{}, ErrInvalidUUID
	}
	return parsed, nil
}

func mapNotificationToModel(n repository.Notification) model.Notification {
	contextData := map[string]any{}
	if len(n.Context) > 0 {
		if err := json.Unmarshal(n.Context, &contextData); err != nil {
			contextData = map[string]any{}
		}
	}

	subjectID := utils.UUIDToString(n.SubjectID)
	content, action := notification.BuildPresentation(n.EventType, subjectID, contextData)

	return model.Notification{
		ID:        utils.UUIDToString(n.ID),
		UserID:    utils.UUIDToString(n.UserID),
		EventType: n.EventType,
		Subject: model.NotificationSubject{
			Type: n.SubjectType,
			ID:   subjectID,
		},
		Context:   contextData,
		Content:   content,
		Action:    action,
		IsRead:    n.ReadAt.Valid,
		CreatedAt: n.CreatedAt.Time,
	}
}
