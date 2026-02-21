package middleware

import (
	"context"
	"fmt"
	"net/http"

	"hr-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

// HRQuerier defines the interface for HR role checks
type HRQuerier interface {
	GetEmployeeByUserID(ctx context.Context, userID pgtype.UUID) (repository.Employee, error)
}

// QueriesAdapter adapts repository.Querier to work with middleware
type QueriesAdapter struct {
	q repository.Querier
}

// NewQueriesAdapter creates a new adapter from a Querier
func NewQueriesAdapter(q repository.Querier) *QueriesAdapter {
	return &QueriesAdapter{q: q}
}

// GetEmployeeByUserID delegates to the underlying querier
func (a *QueriesAdapter) GetEmployeeByUserID(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
	return a.q.GetEmployeeByUserID(ctx, userID)
}

// CheckIsAdmin delegates to the underlying querier
func (a *QueriesAdapter) CheckIsAdmin(ctx context.Context, id pgtype.UUID) (bool, error) {
	return a.q.CheckIsAdmin(ctx, id)
}

// CheckRecruiterRole delegates to the underlying querier
func (a *QueriesAdapter) CheckRecruiterRole(ctx context.Context, employeeID pgtype.UUID) (pgtype.UUID, error) {
	return a.q.CheckRecruiterRole(ctx, employeeID)
}

// GetActiveInterviewCount delegates to the underlying querier
func (a *QueriesAdapter) GetActiveInterviewCount(ctx context.Context, interviewerID pgtype.UUID) (int64, error) {
	return a.q.GetActiveInterviewCount(ctx, interviewerID)
}

func (a *QueriesAdapter) IsCandidateReviewer(ctx context.Context, arg repository.IsCandidateReviewerParams) (pgtype.UUID, error) {
	return a.q.IsCandidateReviewer(ctx, arg)
}

func (a *QueriesAdapter) GetEmployee(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
	return a.q.GetEmployee(ctx, id)
}

func (a *QueriesAdapter) CountCandidateReviewerAssignments(ctx context.Context, reviewerID pgtype.UUID) (int64, error) {
	return a.q.CountCandidateReviewerAssignments(ctx, reviewerID)
}

// RequireAdmin middleware checks if the current user is an admin
func RequireAdmin(queries *repository.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userID, err := parseUUID(userIDStr.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		isAdmin, err := queries.CheckIsAdmin(c.Request.Context(), userID)
		if err != nil || !isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}

		c.Next()
	}
}

// RequireRecruiter middleware checks if the current user is a recruiter
func RequireRecruiter(queries *repository.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userID, err := parseUUID(userIDStr.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		ctx := c.Request.Context()

		isAdmin, err := queries.CheckIsAdmin(ctx, userID)
		if err == nil && isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin account is isolated from recruiter endpoints"})
			return
		}

		// Get employee by user ID
		employee, err := queries.GetEmployeeByUserID(ctx, userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Current user has no linked employee profile"})
			return
		}

		// Check if recruiter
		recruiterID, err := queries.CheckRecruiterRole(ctx, employee.ID)
		if err != nil || !recruiterID.Valid {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Recruiter access required"})
			return
		}

		// Store employee ID in context for later use
		c.Set("employeeID", uuidToString(employee.ID))
		c.Next()
	}
}

// RequireRecruiterOrAdmin middleware allows admin or recruiter access
func RequireRecruiterOrAdmin(queries *repository.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userID, err := parseUUID(userIDStr.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		ctx := c.Request.Context()

		isAdmin, err := queries.CheckIsAdmin(ctx, userID)
		if err == nil && isAdmin {
			c.Next()
			return
		}

		employee, err := queries.GetEmployeeByUserID(ctx, userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Current user has no linked employee profile"})
			return
		}

		recruiterID, err := queries.CheckRecruiterRole(ctx, employee.ID)
		if err != nil || !recruiterID.Valid {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Recruiter access required"})
			return
		}

		c.Set("employeeID", uuidToString(employee.ID))
		c.Next()
	}
}

// RequireHR middleware checks if the current user is an HR employee
func RequireHR(queries *QueriesAdapter) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userID, err := parseUUID(userIDStr.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		ctx := c.Request.Context()

		isAdmin, err := queries.CheckIsAdmin(ctx, userID)
		if err == nil && isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin account is isolated from HR endpoints"})
			return
		}

		// Get employee by user ID
		employee, err := queries.GetEmployeeByUserID(ctx, userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Current user has no linked employee profile"})
			return
		}

		// Check if HR
		if employee.EmployeeType != "HR" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "HR access required"})
			return
		}

		// Store employee ID in context for later use
		c.Set("employeeID", uuidToString(employee.ID))
		c.Next()
	}
}

// RequireInterviewerOrRecruiter middleware checks whether the current user can access interview endpoints.
func RequireInterviewerOrRecruiter(queries *QueriesAdapter) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr, exists := c.Get("userID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}

		userID, err := parseUUID(userIDStr.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}

		ctx := c.Request.Context()

		// Strict isolation: admins cannot access interviewer/recruiter endpoints.
		isAdmin, err := queries.CheckIsAdmin(ctx, userID)
		if err == nil && isAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin account is isolated from interviewer endpoints"})
			return
		}

		employee, err := queries.GetEmployeeByUserID(ctx, userID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Current user has no linked employee profile"})
			return
		}

		// Recruiters can access interview resources for assignment and coordination.
		if recruiterID, err := queries.CheckRecruiterRole(ctx, employee.ID); err == nil && recruiterID.Valid {
			c.Set("employeeID", uuidToString(employee.ID))
			c.Next()
			return
		}

		// Check if explicit interviewer
		interviewerID, err := queries.q.CheckInterviewerRole(ctx, employee.ID)
		if err == nil && interviewerID.Valid {
			c.Set("employeeID", uuidToString(employee.ID))
			c.Next()
			return
		}

		activeInterviewCount, err := queries.GetActiveInterviewCount(ctx, employee.ID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify interviewer access"})
			return
		}
		if activeInterviewCount > 0 {
			c.Set("employeeID", uuidToString(employee.ID))
			c.Next()
			return
		}

		reviewerCount, err := queries.CountCandidateReviewerAssignments(ctx, employee.ID)
		if err == nil && reviewerCount > 0 {
			c.Set("employeeID", uuidToString(employee.ID))
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Interviewer access required"})
	}
}

func RequireCandidateReviewer(queries *QueriesAdapter) gin.HandlerFunc {
	return func(c *gin.Context) {
		employeeIDStr, exists := c.Get("employeeID")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Employee ID not found in context"})
			return
		}

		employeeID, err := parseUUID(employeeIDStr.(string))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
			return
		}

		candidateIDStr := c.Param("id")
		if candidateIDStr == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Candidate ID not found in URL"})
			return
		}

		candidateID, err := parseUUID(candidateIDStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate ID"})
			return
		}

		ctx := c.Request.Context()

		// A user can review a candidate if:
		// 1. They are a Recruiter (broad access)
		// 2. They are explicitly assigned to review this specific candidate

		recruiterID, err := queries.CheckRecruiterRole(ctx, employeeID)
		if err == nil && recruiterID.Valid {
			c.Next()
			return
		}

		_, err = queries.IsCandidateReviewer(ctx, repository.IsCandidateReviewerParams{
			CandidateID: candidateID,
			ReviewerID:  employeeID,
		})
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "You are not assigned to review this candidate"})
			return
		}

		c.Next()
	}
}

func parseUUID(s string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	err := uuid.Scan(s)
	return uuid, err
}

func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	src := u.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", src[0:4], src[4:6], src[6:8], src[8:10], src[10:16])
}
