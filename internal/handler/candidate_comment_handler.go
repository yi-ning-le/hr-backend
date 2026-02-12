package handler

import (
	"errors"
	"net/http"

	"hr-backend/internal/model"
	"hr-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type CandidateCommentHandler struct {
	service *service.CandidateCommentService
}

func NewCandidateCommentHandler(s *service.CandidateCommentService) *CandidateCommentHandler {
	return &CandidateCommentHandler{service: s}
}

func (h *CandidateCommentHandler) ListComments(c *gin.Context) {
	id := candidateIDFromParams(c)
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "candidate ID is required"})
		return
	}
	comments, err := h.service.ListComments(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, comments)
}

func (h *CandidateCommentHandler) CreateComment(c *gin.Context) {
	id := candidateIDFromParams(c)
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "candidate ID is required"})
		return
	}

	var input model.CreateCommentInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employeeID, exists := c.Get("employeeID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: no linked employee profile"})
		return
	}

	comment, err := h.service.CreateComment(c.Request.Context(), id, employeeID.(string), input.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, comment)
}

func (h *CandidateCommentHandler) DeleteComment(c *gin.Context) {
	commentID := c.Param("commentId")

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	employeeID, exists := c.Get("employeeID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: no linked employee profile"})
		return
	}

	err := h.service.DeleteComment(c.Request.Context(), commentID, userID.(string), employeeID.(string))
	if err != nil {
		if errors.Is(err, service.ErrCommentNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		if errors.Is(err, service.ErrDeleteCommentNoPerm) {
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func candidateIDFromParams(c *gin.Context) string {
	id := c.Param("id")
	if id != "" {
		return id
	}
	return c.Param("candidateId")
}
