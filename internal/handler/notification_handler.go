package handler

import (
	"net/http"
	"strconv"

	"hr-backend/internal/model"
	"hr-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type NotificationHandler struct {
	service *service.NotificationService
}

func NewNotificationHandler(s *service.NotificationService) *NotificationHandler {
	return &NotificationHandler{service: s}
}

// GetUserNotifications godoc
func (h *NotificationHandler) GetUserNotifications(c *gin.Context) {
	// Assume user ID is extracted via Auth middleware
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	limitStr := c.Query("limit")
	offsetStr := c.Query("offset")
	var limit, offset int32 = 50, 0

	if l, err := strconv.ParseInt(limitStr, 10, 32); err == nil && l > 0 {
		limit = int32(l)
	}
	if o, err := strconv.ParseInt(offsetStr, 10, 32); err == nil {
		offset = int32(o)
	}

	notifications, err := h.service.GetUserNotifications(c.Request.Context(), userID.(string), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, notifications)
}

// GetUnreadCount godoc
func (h *NotificationHandler) GetUnreadCount(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	count, err := h.service.GetUnreadCount(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, model.NotificationUnreadCount{Count: count})
}

// MarkAsRead godoc
func (h *NotificationHandler) MarkAsRead(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := c.Param("id")
	err := h.service.MarkAsRead(c.Request.Context(), id, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Send something back even if 204 is optimal, consistent with frontend needs
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// MarkAllAsRead godoc
func (h *NotificationHandler) MarkAllAsRead(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	err := h.service.MarkAllAsRead(c.Request.Context(), userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
