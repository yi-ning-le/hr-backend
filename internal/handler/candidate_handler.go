package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"hr-backend/internal/model"
	"hr-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CandidateHandler struct {
	service *service.CandidateService
}

var allowedReviewStatuses = map[string]struct{}{
	"pending":    {},
	"suitable":   {},
	"unsuitable": {},
}

func NewCandidateHandler(s *service.CandidateService) *CandidateHandler {
	return &CandidateHandler{service: s}
}

func (h *CandidateHandler) ListCandidates(c *gin.Context) {
	jobID := c.Query("jobId")
	reviewerID := c.Query("reviewerId")
	reviewStatus := c.Query("reviewStatus")
	search := c.Query("q")
	status := c.Query("status")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	candidates, total, err := h.service.ListCandidates(c.Request.Context(), jobID, reviewerID, reviewStatus, status, search, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": candidates,
		"meta": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
		},
	})
}

func (h *CandidateHandler) GetCandidateCounts(c *gin.Context) {
	counts, err := h.service.GetCandidateCountsByJob(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, counts)
}

func (h *CandidateHandler) CreateCandidate(c *gin.Context) {
	var input model.CandidateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	candidate, err := h.service.CreateCandidate(c.Request.Context(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, candidate)
}

func (h *CandidateHandler) GetCandidate(c *gin.Context) {
	id := c.Param("id")
	candidate, err := h.service.GetCandidate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, candidate)
}

func (h *CandidateHandler) UpdateCandidate(c *gin.Context) {
	id := c.Param("id")
	var input model.CandidateInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	candidate, err := h.service.UpdateCandidate(c.Request.Context(), id, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, candidate)
}

func (h *CandidateHandler) AssignReviewer(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ReviewerID string `json:"reviewerId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	candidate, err := h.service.AssignReviewer(c.Request.Context(), id, req.ReviewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, candidate)
}

func (h *CandidateHandler) SubmitReview(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		ReviewStatus string `json:"reviewStatus" binding:"required"`
		ReviewNote   string `json:"reviewNote"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	reviewStatus := strings.ToLower(strings.TrimSpace(req.ReviewStatus))
	if _, ok := allowedReviewStatuses[reviewStatus]; !ok {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "reviewStatus must be one of: pending, suitable, unsuitable",
		})
		return
	}

	candidate, err := h.service.SubmitReview(c.Request.Context(), id, reviewStatus, req.ReviewNote)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, candidate)
}

func (h *CandidateHandler) DeleteCandidate(c *gin.Context) {
	id := c.Param("id")
	if err := h.service.DeleteCandidate(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *CandidateHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	candidate, err := h.service.UpdateStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, candidate)
}

func (h *CandidateHandler) UpdateNote(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Note string `json:"note" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	candidate, err := h.service.UpdateNote(c.Request.Context(), id, req.Note)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, candidate)
}

func (h *CandidateHandler) UploadResume(c *gin.Context) {
	id := c.Param("id")
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file is received"})
		return
	}

	// Ensure uploads directory exists
	uploadDir := "./uploads"
	if _, err := os.Stat(uploadDir); os.IsNotExist(err) {
		os.Mkdir(uploadDir, 0755)
	}

	// Generate filename
	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + ext
	filePath := filepath.Join(uploadDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}

	// Create URL (assuming static file serving at /static/resumes)
	// We need to return the full URL or relative path.
	// For now, let's return a relative URL that the frontend can prepend the base URL to, or an absolute URL if we knew the host.
	// We'll use relative: /static/resumes/filename
	resumeUrl := fmt.Sprintf("/static/resumes/%s", filename)

	// Update candidate resume URL in DB
	// The OpenAPI spec says this returns { resumeUrl: string, candidate: Candidate }
	// We need to call service to update the URL

	candidate, err := h.service.UpdateResume(c.Request.Context(), id, resumeUrl)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"resumeUrl": resumeUrl,
		"candidate": candidate,
	})
}
