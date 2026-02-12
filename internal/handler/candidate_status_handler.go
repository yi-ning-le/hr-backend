package handler

import (
	"errors"
	"net/http"

	"hr-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type CandidateStatusHandler struct {
	service *service.CandidateStatusService
}

func NewCandidateStatusHandler(s *service.CandidateStatusService) *CandidateStatusHandler {
	return &CandidateStatusHandler{service: s}
}

func (h *CandidateStatusHandler) ListStatuses(c *gin.Context) {
	statuses, err := h.service.ListStatuses(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, statuses)
}

func (h *CandidateStatusHandler) CreateStatus(c *gin.Context) {
	var req struct {
		Name  string `json:"name" binding:"required"`
		Color string `json:"color" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, err := h.service.CreateStatus(c.Request.Context(), req.Name, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, status)
}

func (h *CandidateStatusHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Name  string `json:"name" binding:"required"`
		Color string `json:"color" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status, err := h.service.UpdateStatus(c.Request.Context(), id, req.Name, req.Color)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, status)
}

func (h *CandidateStatusHandler) DeleteStatus(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteStatus(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *CandidateStatusHandler) ReorderStatuses(c *gin.Context) {
	var req struct {
		IDs []string `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.service.ReorderStatuses(c.Request.Context(), req.IDs); err != nil {
		if errors.Is(err, service.ErrInvalidStatusID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusOK)
}
