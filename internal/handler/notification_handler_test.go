package handler_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hr-backend/internal/handler"
	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/service"
	"hr-backend/internal/utils"
	"hr-backend/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func setupTestContext(userID string) (*gin.Engine, *gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, r := gin.CreateTestContext(w)

	r.Use(func(ctx *gin.Context) {
		if userID != "" {
			ctx.Set("userID", userID)
		}
		ctx.Next()
	})

	req, _ := http.NewRequest("GET", "/", nil)
	c.Request = req
	return r, c, w
}

func TestGetUserNotifications(t *testing.T) {
	userIDStr := "e3b0c442-989b-464c-32d8-c19a5c8b66e1"
	userID, _ := utils.StringToUUID(userIDStr)
	candidateID, _ := utils.StringToUUID("11111111-1111-1111-1111-111111111111")
	now := time.Now()
	contextJSON, _ := json.Marshal(map[string]any{
		"candidateId": "11111111-1111-1111-1111-111111111111",
	})

	mockRepo := &mocks.MockQuerier{
		GetNotificationsByUserIdFunc: func(ctx context.Context, arg repository.GetNotificationsByUserIdParams) ([]repository.Notification, error) {
			return []repository.Notification{
				{
					ID:          pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
					UserID:      userID,
					EventType:   model.NotificationEventCandidateReviewerAssigned,
					SubjectType: model.NotificationSubjectTypeCandidate,
					SubjectID:   candidateID,
					Context:     contextJSON,
					ReadAt:      pgtype.Timestamptz{},
					CreatedAt:   pgtype.Timestamptz{Time: now, Valid: true},
				},
			}, nil
		},
	}

	svc := service.NewNotificationService(mockRepo)
	h := handler.NewNotificationHandler(svc)

	r, _, w := setupTestContext(userIDStr)
	r.GET("/notifications", h.GetUserNotifications)

	req, _ := http.NewRequest("GET", "/notifications?limit=10&offset=0", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var notifications []model.Notification
	err := json.Unmarshal(w.Body.Bytes(), &notifications)
	assert.NoError(t, err)
	assert.Len(t, notifications, 1)
	assert.Equal(t, model.NotificationEventCandidateReviewerAssigned, notifications[0].EventType)
	assert.Equal(t, model.NotificationSubjectTypeCandidate, notifications[0].Subject.Type)
	assert.Equal(t, "11111111-1111-1111-1111-111111111111", notifications[0].Subject.ID)
	assert.Equal(t, false, notifications[0].IsRead)
}

func TestGetUnreadCount(t *testing.T) {
	userIDStr := "e3b0c442-989b-464c-32d8-c19a5c8b66e1"

	mockRepo := &mocks.MockQuerier{
		GetUnreadNotificationCountFunc: func(ctx context.Context, userID pgtype.UUID) (int64, error) {
			return 3, nil
		},
	}

	svc := service.NewNotificationService(mockRepo)
	h := handler.NewNotificationHandler(svc)

	r, _, w := setupTestContext(userIDStr)
	r.GET("/notifications/unread-count", h.GetUnreadCount)

	req, _ := http.NewRequest("GET", "/notifications/unread-count", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var result struct {
		Count int64 `json:"count"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.Equal(t, int64(3), result.Count)
}

func TestMarkAsRead(t *testing.T) {
	userIDStr := "e3b0c442-989b-464c-32d8-c19a5c8b66e1"
	notificationIDStr := "f4b0c442-989b-464c-32d8-c19a5c8b66e2"

	called := false
	mockRepo := &mocks.MockQuerier{
		MarkNotificationAsReadFunc: func(ctx context.Context, arg repository.MarkNotificationAsReadParams) error {
			called = true
			return nil
		},
	}

	svc := service.NewNotificationService(mockRepo)
	h := handler.NewNotificationHandler(svc)

	r, _, w := setupTestContext(userIDStr)
	r.PUT("/notifications/:id/read", h.MarkAsRead)

	req, _ := http.NewRequest("PUT", "/notifications/"+notificationIDStr+"/read", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, called)
}

func TestMarkAsRead_InvalidID(t *testing.T) {
	userIDStr := "e3b0c442-989b-464c-32d8-c19a5c8b66e1"
	mockRepo := &mocks.MockQuerier{}

	svc := service.NewNotificationService(mockRepo)
	h := handler.NewNotificationHandler(svc)

	r, _, w := setupTestContext(userIDStr)
	r.PUT("/notifications/:id/read", h.MarkAsRead)

	req, _ := http.NewRequest("PUT", "/notifications/not-a-uuid/read", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMarkAllAsRead(t *testing.T) {
	userIDStr := "e3b0c442-989b-464c-32d8-c19a5c8b66e1"

	called := false
	mockRepo := &mocks.MockQuerier{
		MarkAllNotificationsAsReadFunc: func(ctx context.Context, userID pgtype.UUID) error {
			called = true
			return nil
		},
	}

	svc := service.NewNotificationService(mockRepo)
	h := handler.NewNotificationHandler(svc)

	r, _, w := setupTestContext(userIDStr)
	r.PUT("/notifications/read-all", h.MarkAllAsRead)

	req, _ := http.NewRequest("PUT", "/notifications/read-all", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, called)
}

func TestDeleteNotification(t *testing.T) {
	userIDStr := "e3b0c442-989b-464c-32d8-c19a5c8b66e1"
	notificationIDStr := "f4b0c442-989b-464c-32d8-c19a5c8b66e2"

	called := false
	var capturedArg repository.DeleteNotificationParams
	mockRepo := &mocks.MockQuerier{
		DeleteNotificationFunc: func(ctx context.Context, arg repository.DeleteNotificationParams) error {
			called = true
			capturedArg = arg
			return nil
		},
	}

	svc := service.NewNotificationService(mockRepo)
	h := handler.NewNotificationHandler(svc)

	r, _, w := setupTestContext(userIDStr)
	r.DELETE("/notifications/:id", h.DeleteNotification)

	req, _ := http.NewRequest("DELETE", "/notifications/"+notificationIDStr, nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, called)

	expectedID, _ := utils.StringToUUID(notificationIDStr)
	expectedUserID, _ := utils.StringToUUID(userIDStr)
	assert.Equal(t, expectedID, capturedArg.ID)
	assert.Equal(t, expectedUserID, capturedArg.UserID)
}
