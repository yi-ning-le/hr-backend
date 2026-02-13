package handler

import (
	"context"
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
		// User has no employee record.
		c.JSON(http.StatusOK, model.RecruitmentRoleResponse{
			IsAdmin:          isAdmin,
			IsRecruiter:      false,
			IsInterviewer:    false,
			IsHR:             false,
			CanReviewResumes: false,
		})
		return
	}

	// Check if recruiter
	recruiterID, err := h.queries.CheckRecruiterRole(ctx, employee.ID)
	isRecruiter := err == nil && recruiterID.Valid

	canReviewResumes := employee.CanReviewResumes

	reviewerCount, err := h.queries.CountCandidateReviewerAssignments(ctx, employee.ID)
	if err == nil && reviewerCount > 0 {
		canReviewResumes = true
	}

	// Check if HR
	isHR := employee.EmployeeType == "HR"

	c.JSON(http.StatusOK, model.RecruitmentRoleResponse{
		IsAdmin:          isAdmin,
		IsRecruiter:      isRecruiter,
		IsInterviewer:    canReviewResumes, // Backward compatible alias.
		IsHR:             isHR,
		CanReviewResumes: canReviewResumes,
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

// GetHRs lists all HR employees (Admin only)
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

// AssignHR assigns HR role to an employee (Admin only)
func (h *RecruitmentHandler) AssignHR(c *gin.Context) {
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
	if err := h.queries.AssignHRRole(ctx, employeeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign HR role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "HR role assigned"})
}

// RevokeHR removes HR role from an employee (Admin only)
func (h *RecruitmentHandler) RevokeHR(c *gin.Context) {
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
	if err := h.queries.RevokeHRRole(ctx, employeeID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke HR role"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "HR role revoked"})
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

// CreateInterview assigns a candidate to an interviewer (creates an interview)
func (h *RecruitmentHandler) CreateInterview(c *gin.Context) {
	var input model.CreateInterviewInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	candidateID, err := parseUUID(input.CandidateID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid candidate ID"})
		return
	}

	interviewerID, err := parseUUID(input.InterviewerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interviewer ID"})
		return
	}

	jobID, err := parseUUID(input.JobID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid job ID"})
		return
	}

	ctx := c.Request.Context()
	// Create interview
	interview, err := h.queries.CreateInterview(ctx, repository.CreateInterviewParams{
		CandidateID:   candidateID,
		InterviewerID: interviewerID,
		JobID:         jobID,
		ScheduledTime: pgtype.Timestamptz{Time: input.ScheduledTime, Valid: true},
		Status:        "PENDING",
		Notes:         pgtype.Text{String: input.Notes, Valid: input.Notes != ""},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create interview"})
		return
	}

	c.JSON(http.StatusCreated, model.Interview{
		ID:            uuidToString(interview.ID),
		CandidateID:   uuidToString(interview.CandidateID),
		InterviewerID: uuidToString(interview.InterviewerID),
		JobID:         uuidToString(interview.JobID),
		ScheduledTime: interview.ScheduledTime.Time,
		Status:        interview.Status,
		Notes:         interview.Notes.String,
		CreatedAt:     interview.CreatedAt.Time,
	})
}

// GetMyInterviews returns interviews assigned to the current employee
func (h *RecruitmentHandler) GetMyInterviews(c *gin.Context) {
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

	// Get employee ID
	employee, err := h.queries.GetEmployeeByUserID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Employee profile not found"})
		return
	}

	// List interviews
	interviews, err := h.queries.ListInterviewsByInterviewer(ctx, employee.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list interviews"})
		return
	}

	result := make([]model.Interview, len(interviews))
	for i, interview := range interviews {
		result[i] = model.Interview{
			ID:            uuidToString(interview.ID),
			CandidateID:   uuidToString(interview.CandidateID),
			InterviewerID: uuidToString(interview.InterviewerID),
			JobID:         uuidToString(interview.JobID),
			ScheduledTime: interview.ScheduledTime.Time,
			Status:        interview.Status,
			Notes:         interview.Notes.String,
			CreatedAt:     interview.CreatedAt.Time,
		}
	}

	c.JSON(http.StatusOK, result)
}

// GetInterview returns a specific interview detail
func (h *RecruitmentHandler) GetInterview(c *gin.Context) {
	interviewIDStr := c.Param("id")
	interviewID, err := parseUUID(interviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}

	userID, ok := currentUserIDFromContext(c)
	if !ok {
		return
	}

	ctx := c.Request.Context()
	interview, err := h.queries.GetInterview(ctx, interviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
		return
	}

	canAccess, err := h.canAccessInterviewByInterviewer(ctx, userID, interview.InterviewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify interview permission"})
		return
	}
	if !canAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Interview access denied"})
		return
	}

	c.JSON(http.StatusOK, model.Interview{
		ID:            uuidToString(interview.ID),
		CandidateID:   uuidToString(interview.CandidateID),
		InterviewerID: uuidToString(interview.InterviewerID),
		JobID:         uuidToString(interview.JobID),
		ScheduledTime: interview.ScheduledTime.Time,
		Status:        interview.Status,
		Notes:         interview.Notes.String,
		CreatedAt:     interview.CreatedAt.Time,
	})
}

// UpdateInterviewNotes updates the notes for an interview
func (h *RecruitmentHandler) UpdateInterviewNotes(c *gin.Context) {
	interviewIDStr := c.Param("id")
	interviewID, err := parseUUID(interviewIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interview ID"})
		return
	}

	userID, ok := currentUserIDFromContext(c)
	if !ok {
		return
	}

	var input model.UpdateInterviewNotesInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()
	existingInterview, err := h.queries.GetInterview(ctx, interviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
		return
	}

	canAccess, err := h.canAccessInterviewByInterviewer(
		ctx,
		userID,
		existingInterview.InterviewerID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify interview permission"})
		return
	}
	if !canAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "Interview access denied"})
		return
	}

	interview, err := h.queries.UpdateInterviewNote(ctx, repository.UpdateInterviewNoteParams{
		ID:    interviewID,
		Notes: pgtype.Text{String: input.Notes, Valid: input.Notes != ""},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update interview notes"})
		return
	}

	c.JSON(http.StatusOK, model.Interview{
		ID:            uuidToString(interview.ID),
		CandidateID:   uuidToString(interview.CandidateID),
		InterviewerID: uuidToString(interview.InterviewerID),
		JobID:         uuidToString(interview.JobID),
		ScheduledTime: interview.ScheduledTime.Time,
		Status:        interview.Status,
		Notes:         interview.Notes.String,
		CreatedAt:     interview.CreatedAt.Time,
	})
}

func currentUserIDFromContext(c *gin.Context) (pgtype.UUID, bool) {
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return pgtype.UUID{}, false
	}

	userID, err := parseUUID(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return pgtype.UUID{}, false
	}

	return userID, true
}

func (h *RecruitmentHandler) canAccessInterviewByInterviewer(
	ctx context.Context,
	userID pgtype.UUID,
	interviewerID pgtype.UUID,
) (bool, error) {
	employee, err := h.queries.GetEmployeeByUserID(ctx, userID)
	if err != nil {
		return false, nil
	}

	if recruiterID, checkErr := h.queries.CheckRecruiterRole(ctx, employee.ID); checkErr == nil && recruiterID.Valid {
		return true, nil
	}

	return employee.ID == interviewerID, nil
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
