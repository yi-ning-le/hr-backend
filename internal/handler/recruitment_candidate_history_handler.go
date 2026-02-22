package handler

import (
	"net/http"
	"strings"
	"time"

	"hr-backend/internal/repository"

	"github.com/gin-gonic/gin"
)

type candidateHistoryItem struct {
	CandidateID   string    `json:"candidateId"`
	CandidateName string    `json:"candidateName"`
	Status        string    `json:"status"`
	ReviewStatus  string    `json:"reviewStatus"`
	AppliedAt     time.Time `json:"appliedAt"`
	JobTitle      string    `json:"jobTitle"`
}

type reviewerCandidateItem struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	ExperienceYears int32     `json:"experienceYears"`
	Education       string    `json:"education"`
	AppliedJobID    string    `json:"appliedJobId"`
	AppliedJobTitle string    `json:"appliedJobTitle"`
	Channel         string    `json:"channel"`
	ResumeURL       string    `json:"resumeUrl"`
	Status          string    `json:"status"`
	ReviewStatus    string    `json:"reviewStatus"`
	AppliedAt       time.Time `json:"appliedAt"`
}

// GetCandidateHistory returns the history of a candidate.
// Default scope is the current reviewer. Recruiters/Admin can request scope=all.
func (h *RecruitmentHandler) GetCandidateHistory(c *gin.Context) {
	candidateIDStr := c.Param("id")
	candidateID, err := parseUUID(candidateIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate ID"})
		return
	}

	userID, ok := currentUserIDFromContext(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	scope := strings.ToLower(strings.TrimSpace(c.DefaultQuery("scope", "self")))

	if scope == "all" {
		isAdmin, _ := h.queries.CheckIsAdmin(ctx, userID)
		if !isAdmin {
			employee, empErr := h.queries.GetEmployeeByUserID(ctx, userID)
			if empErr != nil {
				c.JSON(http.StatusForbidden, gin.H{"error": "Current user has no linked employee profile"})
				return
			}
			recruiterID, roleErr := h.queries.CheckRecruiterRole(ctx, employee.ID)
			if roleErr != nil || !recruiterID.Valid {
				c.JSON(http.StatusForbidden, gin.H{"error": "Recruiter access required for full history"})
				return
			}
		}

		history, historyErr := h.queries.GetCandidateHistory(ctx, candidateID)
		if historyErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get candidate history"})
			return
		}

		c.JSON(http.StatusOK, mapCandidateHistoryRows(history))
		return
	}

	employee, err := h.queries.GetEmployeeByUserID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Current user has no linked employee profile"})
		return
	}

	history, err := h.queries.GetCandidateHistoryForReviewer(ctx, repository.GetCandidateHistoryForReviewerParams{
		CandidateID: candidateID,
		ReviewerID:  employee.ID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get candidate history"})
		return
	}

	c.JSON(http.StatusOK, mapCandidateHistoryForReviewerRows(history))
}

// GetPendingReviewCandidates returns candidates that the current reviewer still needs to review.
func (h *RecruitmentHandler) GetPendingReviewCandidates(c *gin.Context) {
	userID, ok := currentUserIDFromContext(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	employee, err := h.queries.GetEmployeeByUserID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Current user has no linked employee profile"})
		return
	}

	candidates, err := h.queries.ListPendingReviewCandidates(ctx, employee.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get pending review candidates"})
		return
	}

	c.JSON(http.StatusOK, mapPendingReviewCandidateRows(candidates))
}

// GetPastReviewedCandidates returns candidates that the current reviewer has reviewed in the past.
func (h *RecruitmentHandler) GetPastReviewedCandidates(c *gin.Context) {
	userID, ok := currentUserIDFromContext(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	employee, err := h.queries.GetEmployeeByUserID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Current user has no linked employee profile"})
		return
	}

	candidates, err := h.queries.GetPastReviewedCandidates(ctx, employee.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get past reviewed candidates"})
		return
	}

	c.JSON(http.StatusOK, mapPastReviewedCandidateRows(candidates))
}

func mapCandidateHistoryRows(rows []repository.GetCandidateHistoryRow) []candidateHistoryItem {
	items := make([]candidateHistoryItem, len(rows))
	for i, row := range rows {
		items[i] = candidateHistoryItem{
			CandidateID:   uuidToString(row.CandidateID),
			CandidateName: row.CandidateName,
			Status:        row.Status,
			ReviewStatus:  row.ReviewStatus,
			AppliedAt:     row.AppliedAt.Time,
			JobTitle:      row.JobTitle,
		}
	}
	return items
}

func mapCandidateHistoryForReviewerRows(rows []repository.GetCandidateHistoryForReviewerRow) []candidateHistoryItem {
	items := make([]candidateHistoryItem, len(rows))
	for i, row := range rows {
		items[i] = candidateHistoryItem{
			CandidateID:   uuidToString(row.CandidateID),
			CandidateName: row.CandidateName,
			Status:        row.Status,
			ReviewStatus:  row.ReviewStatus,
			AppliedAt:     row.AppliedAt.Time,
			JobTitle:      row.JobTitle,
		}
	}
	return items
}

func mapPendingReviewCandidateRows(rows []repository.ListPendingReviewCandidatesRow) []reviewerCandidateItem {
	items := make([]reviewerCandidateItem, len(rows))
	for i, row := range rows {
		items[i] = reviewerCandidateItem{
			ID:              uuidToString(row.ID),
			Name:            row.Name,
			Email:           row.Email,
			Phone:           row.Phone,
			ExperienceYears: row.ExperienceYears,
			Education:       row.Education,
			AppliedJobID:    uuidToString(row.AppliedJobID),
			AppliedJobTitle: row.AppliedJobTitle,
			Channel:         row.Channel,
			ResumeURL:       row.ResumeUrl,
			Status:          row.Status,
			ReviewStatus:    row.ReviewStatus,
			AppliedAt:       row.AppliedAt.Time,
		}
	}
	return items
}

func mapPastReviewedCandidateRows(rows []repository.GetPastReviewedCandidatesRow) []reviewerCandidateItem {
	items := make([]reviewerCandidateItem, len(rows))
	for i, row := range rows {
		items[i] = reviewerCandidateItem{
			ID:              uuidToString(row.ID),
			Name:            row.Name,
			Email:           row.Email,
			Phone:           row.Phone,
			ExperienceYears: row.ExperienceYears,
			Education:       row.Education,
			AppliedJobID:    uuidToString(row.AppliedJobID),
			AppliedJobTitle: row.AppliedJobTitle,
			Channel:         row.Channel,
			ResumeURL:       row.ResumeUrl,
			Status:          row.Status,
			ReviewStatus:    row.ReviewStatus,
			AppliedAt:       row.AppliedAt.Time,
		}
	}
	return items
}
