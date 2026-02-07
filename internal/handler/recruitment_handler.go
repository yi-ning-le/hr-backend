package handler

import (
	"fmt"
	"net/http"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type RecruitmentHandler struct {
	queries repository.Querier
}

func NewRecruitmentHandler(q repository.Querier) *RecruitmentHandler {
	return &RecruitmentHandler{queries: q}
}

// GetMyRole returns the current user's role status
func (h *RecruitmentHandler) GetMyRole(c *gin.Context) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	userID, err := parseUUID(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	ctx := c.Request.Context()

	// Check if admin
	isAdmin, err := h.queries.CheckIsAdmin(ctx, userID)
	if err != nil {
		isAdmin = false
	}

	// Get employee by user ID
	employee, err := h.queries.GetEmployeeByUserID(ctx, userID)
	if err != nil {
		// User has no employee record, return minimal response
		c.JSON(http.StatusOK, model.RecruitmentRoleResponse{
			IsAdmin:       isAdmin,
			IsRecruiter:   false,
			IsInterviewer: false,
		})
		return
	}

	// Check if recruiter
	_, err = h.queries.CheckRecruiterRole(ctx, employee.ID)
	isRecruiter := err == nil

	// Check active interviews
	interviewCount, err := h.queries.GetActiveInterviewCount(ctx, employee.ID)
	isInterviewer := err == nil && interviewCount > 0

	c.JSON(http.StatusOK, model.RecruitmentRoleResponse{
		IsAdmin:       isAdmin,
		IsRecruiter:   isRecruiter,
		IsInterviewer: isInterviewer,
	})
}

// GetRecruiters lists all recruiters (Admin only)
func (h *RecruitmentHandler) GetRecruiters(c *gin.Context) {
	ctx := c.Request.Context()

	recruiters, err := h.queries.ListRecruiters(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list recruiters"})
		return
	}

	result := make([]model.Recruiter, len(recruiters))
	for i, r := range recruiters {
		result[i] = model.Recruiter{
			EmployeeID: uuidToString(r.ID),
			FirstName:  r.FirstName,
			LastName:   r.LastName,
			Department: r.Department,
		}
	}

	c.JSON(http.StatusOK, result)
}

// AssignRecruiter assigns recruiter role to an employee (Admin only)
func (h *RecruitmentHandler) AssignRecruiter(c *gin.Context) {
	var input struct {
		EmployeeID string `json:"employeeId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employeeID, err := parseUUID(input.EmployeeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	ctx := c.Request.Context()
	if err := h.queries.AssignRecruiterRole(ctx, employeeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign recruiter role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recruiter role assigned"})
}

// RevokeRecruiter removes recruiter role from an employee (Admin only)
func (h *RecruitmentHandler) RevokeRecruiter(c *gin.Context) {
	var input struct {
		EmployeeID string `json:"employeeId" binding:"required"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employeeID, err := parseUUID(input.EmployeeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return
	}

	ctx := c.Request.Context()
	if err := h.queries.RevokeRecruiterRole(ctx, employeeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke recruiter role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recruiter role revoked"})
}

// TransferInterview transfers an interview to another employee (Recruiter only)
func (h *RecruitmentHandler) TransferInterview(c *gin.Context) {
	interviewIDStr := c.Param("id")
	interviewID, err := parseUUID(interviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}

	var input model.TransferInterviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newInterviewerID, err := parseUUID(input.NewInterviewerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interviewer ID"})
		return
	}

	ctx := c.Request.Context()
	_, err = h.queries.TransferInterview(ctx, repository.TransferInterviewParams{
		ID:            interviewID,
		InterviewerID: newInterviewerID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to transfer interview"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Interview transferred"})
}

// Helper functions
func parseUUID(s string) (pgtype.UUID, error) {
	var uuid pgtype.UUID
	err := uuid.Scan(s)
	return uuid, err
}

func uuidToString(u pgtype.UUID) string {
	if !u.Valid {
		return ""
	}
	// Use the utility function for consistent UUID string format
	src := u.Bytes
	return fmt.Sprintf("%x-%x-%x-%x-%x", src[0:4], src[4:6], src[6:8], src[8:10], src[10:16])
}
