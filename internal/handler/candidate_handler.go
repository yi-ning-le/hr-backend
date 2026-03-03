package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"hr-backend/internal/model"
	"hr-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
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

const (
	maxResumeUploadSize = 10 << 20
	resumeUploadDir     = "./uploads"
	pdfFileExtension    = ".pdf"
)

func (h *CandidateHandler) saveResumeFile(c *gin.Context, file *multipart.FileHeader) (string, string, error) {
	if err := os.MkdirAll(resumeUploadDir, 0o755); err != nil {
		return "", "", err
	}

	ext := filepath.Ext(file.Filename)
	filename := uuid.New().String() + ext
	filePath := filepath.Join(resumeUploadDir, filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		return "", "", err
	}

	resumeURL := fmt.Sprintf("/static/resumes/%s", filename)
	return filePath, resumeURL, nil
}

func validateResumeFile(file *multipart.FileHeader) error {
	if file.Size <= 0 {
		return errors.New("resume file is empty")
	}
	if file.Size > maxResumeUploadSize {
		return fmt.Errorf("resume file exceeds %dMB limit", maxResumeUploadSize>>20)
	}

	ext := strings.ToLower(filepath.Ext(file.Filename))
	if ext != pdfFileExtension {
		return errors.New("only PDF files are supported")
	}

	contentType := strings.ToLower(file.Header.Get("Content-Type"))
	if contentType != "" &&
		contentType != "application/octet-stream" &&
		!strings.Contains(contentType, "application/pdf") {
		return errors.New("invalid resume content type")
	}

	return nil
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
	// Parse multipart form
	if err := c.Request.ParseMultipartForm(maxResumeUploadSize); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse multipart form"})
		return
	}

	// 1. Extract the file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Resume file is required"})
		return
	}

	// 2. Extract the JSON data
	dataStr := c.PostForm("data")
	if dataStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Candidate data is required"})
		return
	}

	var input model.CandidateCreateInput
	if err := json.Unmarshal([]byte(dataStr), &input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid candidate data format: %v", err)})
		return
	}

	if err := validateResumeFile(file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filePath, resumeURL, err := h.saveResumeFile(c, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save resume file"})
		return
	}

	if err := binding.Validator.ValidateStruct(input); err != nil {
		_ = os.Remove(filePath)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	candidate, err := h.service.CreateCandidate(c.Request.Context(), input, resumeURL)
	if err != nil {
		_ = os.Remove(filePath)
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

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	candidate, err := h.service.AssignReviewer(c.Request.Context(), id, req.ReviewerID, userIDStr)
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
		Comment      string `json:"comment"`
		ReviewNote   string `json:"reviewNote"` // Legacy alias for backward compatibility.
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

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userIDStr, ok := userID.(string)
	if !ok || userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	comment := strings.TrimSpace(req.Comment)
	if comment == "" {
		comment = strings.TrimSpace(req.ReviewNote)
	}

	candidate, err := h.service.SubmitReview(
		c.Request.Context(),
		id,
		userIDStr,
		reviewStatus,
		comment,
	)
	if err != nil {
		if errors.Is(err, service.ErrReviewPermissionDenied) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Only assigned reviewer can submit review"})
			return
		}
		if errors.Is(err, service.ErrReviewerProfileNotFound) {
			c.JSON(http.StatusForbidden, gin.H{"error": "Current user has no linked employee profile"})
			return
		}
		if errors.Is(err, service.ErrCandidateNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Candidate not found"})
			return
		}
		// Keep legacy fallback for older service implementations.
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Candidate not found"})
			return
		}
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

func (h *CandidateHandler) RevertReviewer(c *gin.Context) {
	id := c.Param("id")

	candidate, err := h.service.RevertReviewer(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCandidateID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate ID"})
			return
		}
		if errors.Is(err, service.ErrNoReviewerToRevert) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No reviewer to revert"})
			return
		}
		if errors.Is(err, service.ErrReviewAlreadySubmitted) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Review already submitted, cannot revert"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, candidate)
}
