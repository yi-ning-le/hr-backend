package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"hr-backend/internal/model"
	"hr-backend/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CandidateHandler struct {
	service *service.CandidateService
}

func NewCandidateHandler(s *service.CandidateService) *CandidateHandler {
	return &CandidateHandler{service: s}
}

func (h *CandidateHandler) ListCandidates(c *gin.Context) {
	jobID := c.Query("jobId")
	reviewerID := c.Query("reviewerId")
	reviewStatus := c.Query("reviewStatus")
	search := c.Query("q")
	status := c.Query("status") // not used in service ListCandidates yet? Wait service uses reviewStatus, but frontend passes status?
	// Check service implementation:
	// func (s *CandidateService) ListCandidates(..., reviewStatusFilter string, ...)
	// unique status filter is for candidate status (new, screening, etc), reviewStatus is for review (pending, approved).
	// The previous implementation didn't filter by candidate status in `ListCandidates`?
	// Let's check query.sql:
	// WHERE ... AND (sqlc.narg('review_status')::text IS NULL OR c.review_status = sqlc.narg('review_status'))
	// It seems `status` column on candidates was not filtered in previous ListCandidates?
	// Wait, let me check query.sql again.
	// Yes, previous query.sql only filtered by review_status.
	// But frontend `CandidateManagement` filters by `status`.
	// The previous frontend implementation fetched ALL and filtered locally.
	// So I need to add `status` filter to backend if I want server-side filtering.
	// I missed adding `status` to `ListCandidates` query in previous step!
	// I added `search`, `page`, `limit`.
	// I should probably add `status` (candidate status) too.

	// For now, let's just use what I added. Frontend will need to pass `reviewStatus` if it means that, or I need to add `status` filter.
	// The frontend `CandidateManagement` uses `status` (e.g. 'new', 'screening').
	// So I definitely see a missing piece in my plan. I need to add `status` filter to SQL.

	// Let's finish the handler update for what I have, and then I might need to cycle back to SQL if I want full server-side filtering.
	// Actually, `CandidateManagement.tsx` has `const statusFilter = search.status || [];`
	// This refers to candidate status (new, screening), not review status.
	// The database has `status` column.
	// The `ListCandidates` query I modified:
	// WHERE ... AND (sqlc.narg('review_status')::text IS NULL OR c.review_status = sqlc.narg('review_status'))
	// It does NOT have `c.status = ...`.

	// ACTION: I need to start a sub-task to add `status` filter to SQL.
	// But first, let's minimally fix the build by updating the handler to match current service signature.

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

	candidate, err := h.service.SubmitReview(c.Request.Context(), id, req.ReviewStatus, req.ReviewNote)
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
