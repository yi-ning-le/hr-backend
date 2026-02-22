package handler

import (
	"net/http"

	"hr-backend/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

type employeeIDRequest struct {
	EmployeeID string `json:"employeeId" binding:"required"`
}

// GetMyRole returns the current user's role status.
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
	isAdmin, err := h.queries.CheckIsAdmin(ctx, userID)
	if err != nil {
		isAdmin = false
	}

	employee, err := h.queries.GetEmployeeByUserID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusOK, model.RecruitmentRoleResponse{
			IsAdmin:       isAdmin,
			IsRecruiter:   false,
			IsInterviewer: false,
			IsHR:          false,
		})
		return
	}

	recruiterID, err := h.queries.CheckRecruiterRole(ctx, employee.ID)
	isRecruiter := err == nil && recruiterID.Valid

	interviewerID, interviewerRoleErr := h.queries.CheckInterviewerRole(ctx, employee.ID)
	isInterviewer := interviewerRoleErr == nil && interviewerID.Valid

	c.JSON(http.StatusOK, model.RecruitmentRoleResponse{
		IsAdmin:       isAdmin,
		IsRecruiter:   isRecruiter,
		IsInterviewer: isInterviewer,
		IsHR:          employee.EmployeeType == "HR",
	})
}

// GetRecruiters lists all recruiters (Admin only).
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

// AssignRecruiter assigns recruiter role to an employee (Admin only).
func (h *RecruitmentHandler) AssignRecruiter(c *gin.Context) {
	employeeID, ok := parseEmployeeIDFromBody(c)
	if !ok {
		return
	}

	if err := h.queries.AssignRecruiterRole(c.Request.Context(), employeeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign recruiter role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recruiter role assigned"})
}

// RevokeRecruiter removes recruiter role from an employee (Admin only).
func (h *RecruitmentHandler) RevokeRecruiter(c *gin.Context) {
	employeeID, ok := parseEmployeeIDFromBody(c)
	if !ok {
		return
	}

	if err := h.queries.RevokeRecruiterRole(c.Request.Context(), employeeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke recruiter role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Recruiter role revoked"})
}

// GetInterviewers lists all interviewers (Admin only).
func (h *RecruitmentHandler) GetInterviewers(c *gin.Context) {
	ctx := c.Request.Context()
	interviewers, err := h.queries.ListInterviewers(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list interviewers"})
		return
	}

	result := make([]model.Interviewer, len(interviewers))
	for i, r := range interviewers {
		result[i] = model.Interviewer{
			EmployeeID: uuidToString(r.ID),
			FirstName:  r.FirstName,
			LastName:   r.LastName,
			Department: r.Department,
		}
	}
	c.JSON(http.StatusOK, result)
}

// AssignInterviewer assigns interviewer role to an employee (Admin only).
func (h *RecruitmentHandler) AssignInterviewer(c *gin.Context) {
	employeeID, ok := parseEmployeeIDFromBody(c)
	if !ok {
		return
	}

	if err := h.queries.AssignInterviewerRole(c.Request.Context(), employeeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign interviewer role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Interviewer role assigned"})
}

// RevokeInterviewer removes interviewer role from an employee (Admin only).
func (h *RecruitmentHandler) RevokeInterviewer(c *gin.Context) {
	employeeID, ok := parseEmployeeIDFromBody(c)
	if !ok {
		return
	}

	if err := h.queries.RevokeInterviewerRole(c.Request.Context(), employeeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke interviewer role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Interviewer role revoked"})
}

// GetHRs lists all HR employees (Admin only).
func (h *RecruitmentHandler) GetHRs(c *gin.Context) {
	ctx := c.Request.Context()
	hrs, err := h.queries.ListHRs(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list HRs"})
		return
	}

	result := make([]model.HREmployee, len(hrs))
	for i, hr := range hrs {
		result[i] = model.HREmployee{
			EmployeeID: uuidToString(hr.ID),
			FirstName:  hr.FirstName,
			LastName:   hr.LastName,
			Department: hr.Department,
		}
	}
	c.JSON(http.StatusOK, result)
}

// AssignHR assigns HR role to an employee (Admin only).
func (h *RecruitmentHandler) AssignHR(c *gin.Context) {
	employeeID, ok := parseEmployeeIDFromBody(c)
	if !ok {
		return
	}

	if err := h.queries.AssignHRRole(c.Request.Context(), employeeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign HR role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "HR role assigned"})
}

// RevokeHR removes HR role from an employee (Admin only).
func (h *RecruitmentHandler) RevokeHR(c *gin.Context) {
	employeeID, ok := parseEmployeeIDFromBody(c)
	if !ok {
		return
	}

	if err := h.queries.RevokeHRRole(c.Request.Context(), employeeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke HR role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "HR role revoked"})
}

func parseEmployeeIDFromBody(c *gin.Context) (pgtype.UUID, bool) {
	var input employeeIDRequest
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return pgtype.UUID{}, false
	}

	employeeID, err := parseUUID(input.EmployeeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid employee ID"})
		return pgtype.UUID{}, false
	}

	return employeeID, true
}
