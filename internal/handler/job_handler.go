package handler

import (
	"net/http"

	"hr-backend/internal/model"
	"hr-backend/internal/service"

	"github.com/gin-gonic/gin"
)

type JobHandler struct {
	service *service.JobService
}

func NewJobHandler(s *service.JobService) *JobHandler {
	return &JobHandler{service: s}
}

// ListJobs godoc
func (h *JobHandler) ListJobs(c *gin.Context) {
	jobs, err := h.service.ListJobs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, jobs)
}

// CreateJob godoc
func (h *JobHandler) CreateJob(c *gin.Context) {
	var input model.JobInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job, err := h.service.CreateJob(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, job)
}

// UpdateJob godoc
func (h *JobHandler) UpdateJob(c *gin.Context) {
	id := c.Param("id")
	var input model.JobInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	job, err := h.service.UpdateJob(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, job)
}

// DeleteJob godoc
func (h *JobHandler) DeleteJob(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteJob(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// ToggleStatus godoc
func (h *JobHandler) ToggleStatus(c *gin.Context) {
	id := c.Param("id")
	job, err := h.service.ToggleJobStatus(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, job)
}
