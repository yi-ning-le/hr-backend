package handler

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

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

	activeInterviewCount, activeInterviewErr := h.queries.GetActiveInterviewCount(ctx, employee.ID)
	interviewerID, interviewerRoleErr := h.queries.CheckInterviewerRole(ctx, employee.ID)
	isInterviewer := (interviewerRoleErr == nil && interviewerID.Valid) || (activeInterviewErr == nil && activeInterviewCount > 0)

	c.JSON(http.StatusOK, model.RecruitmentRoleResponse{
		IsAdmin:          isAdmin,
		IsRecruiter:      isRecruiter,
		IsInterviewer:    isInterviewer,
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

	if validationErr := validateInterviewSchedule(input.ScheduledTime, input.ScheduledEndTime, time.Now()); validationErr != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
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

	interview, err := h.queries.CreateInterview(ctx, repository.CreateInterviewParams{
		ID:               candidateID,
		InterviewerID:    interviewerID,
		JobID:            jobID,
		ScheduledTime:    pgtype.Timestamptz{Time: input.ScheduledTime, Valid: true},
		ScheduledEndTime: pgtype.Timestamptz{Time: input.ScheduledEndTime, Valid: true},
		Status:           "PENDING",
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create interview"})
		return
	}

	_ = h.queries.AssignInterviewerRole(ctx, interviewerID)

	var snapshot *model.SnapshotStatus
	if interview.SnapshotStatusKey != "" {
		snapshot = &model.SnapshotStatus{
			Key:   interview.SnapshotStatusKey,
			Label: interview.SnapshotStatusLabel,
		}
	}

	c.JSON(http.StatusCreated, model.Interview{
		ID:               uuidToString(interview.ID),
		CandidateID:      uuidToString(interview.CandidateID),
		InterviewerID:    uuidToString(interview.InterviewerID),
		JobID:            uuidToString(interview.JobID),
		ScheduledTime:    interview.ScheduledTime.Time,
		ScheduledEndTime: interview.ScheduledEndTime.Time,
		Status:           interview.Status,
		CreatedAt:        interview.CreatedAt.Time,
		SnapshotStatus:   snapshot,
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
		var snapshot *model.SnapshotStatus
		if interview.SnapshotStatusKey != "" {
			snapshot = &model.SnapshotStatus{
				Key:   interview.SnapshotStatusKey,
				Label: interview.SnapshotStatusLabel,
			}
		}

		result[i] = model.Interview{
			ID:               uuidToString(interview.ID),
			CandidateID:      uuidToString(interview.CandidateID),
			InterviewerID:    uuidToString(interview.InterviewerID),
			InterviewerName:  employee.FirstName + " " + employee.LastName,
			JobID:            uuidToString(interview.JobID),
			ScheduledTime:    interview.ScheduledTime.Time,
			ScheduledEndTime: interview.ScheduledEndTime.Time,
			Status:           interview.Status,
			CreatedAt:        interview.CreatedAt.Time,
			SnapshotStatus:   snapshot,
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

	var snapshot *model.SnapshotStatus
	if interview.SnapshotStatusKey != "" {
		snapshot = &model.SnapshotStatus{
			Key:   interview.SnapshotStatusKey,
			Label: interview.SnapshotStatusLabel,
		}
	}

	c.JSON(http.StatusOK, model.Interview{
		ID:               uuidToString(interview.ID),
		CandidateID:      uuidToString(interview.CandidateID),
		InterviewerID:    uuidToString(interview.InterviewerID),
		JobID:            uuidToString(interview.JobID),
		ScheduledTime:    interview.ScheduledTime.Time,
		ScheduledEndTime: interview.ScheduledEndTime.Time,
		Status:           interview.Status,
		CreatedAt:        interview.CreatedAt.Time,
		SnapshotStatus:   snapshot,
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

func validateInterviewSchedule(scheduledTime, scheduledEndTime, now time.Time) string {
	if !scheduledTime.After(now) {
		return "Start time must be in the future"
	}

	if !scheduledEndTime.After(scheduledTime) {
		return "End time must be after start time"
	}

	return ""
}

// UpdateInterviewStatus updates the status of an interview
func (h *RecruitmentHandler) UpdateInterviewStatus(c *gin.Context) {
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

	var input model.UpdateInterviewStatusInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := c.Request.Context()

	// 1. Get the interview to check permissions
	interview, err := h.queries.GetInterview(ctx, interviewID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Interview not found"})
		return
	}

	// 2. Check permissions (Recruiter or Assigned Interviewer)
	canAccess, err := h.canAccessInterviewByInterviewer(ctx, userID, interview.InterviewerID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify interview permission"})
		return
	}

	if !canAccess {
		c.JSON(http.StatusForbidden, gin.H{"error": "You do not have permission to update this interview"})
		return
	}

	// 3. Update Status
	updatedInterview, err := h.queries.UpdateInterviewStatus(ctx, repository.UpdateInterviewStatusParams{
		ID:     interviewID,
		Status: input.Status,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update interview status"})
		return
	}

	var snapshot *model.SnapshotStatus
	if interview.SnapshotStatusKey != "" {
		snapshot = &model.SnapshotStatus{
			Key:   interview.SnapshotStatusKey,
			Label: interview.SnapshotStatusLabel,
		}
	}

	c.JSON(http.StatusOK, model.Interview{
		ID:               uuidToString(updatedInterview.ID),
		CandidateID:      uuidToString(updatedInterview.CandidateID),
		InterviewerID:    uuidToString(updatedInterview.InterviewerID),
		JobID:            uuidToString(updatedInterview.JobID),
		ScheduledTime:    updatedInterview.ScheduledTime.Time,
		ScheduledEndTime: updatedInterview.ScheduledEndTime.Time,
		Status:           updatedInterview.Status,
		CreatedAt:        updatedInterview.CreatedAt.Time,
		SnapshotStatus:   snapshot,
	})
}

// GetAllInterviews lists all interviews (Recruiter/Admin only)
func (h *RecruitmentHandler) GetAllInterviews(c *gin.Context) {
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

	hasPermission, err := h.checkRecruiterOrAdmin(ctx, userID)
	if err != nil {
		// Log error internally if needed
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify permissions"})
		return
	}
	if !hasPermission {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	// Parse query params
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("pageSize", "50")
	startStr := c.Query("start")
	endStr := c.Query("end")
	statusesStr := c.Query("status") // Comma separated

	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	pageSize, _ := strconv.Atoi(pageSizeStr)
	if pageSize < 1 {
		pageSize = 50
	}

	limit := int32(pageSize)
	offset := int32((page - 1) * pageSize)

	var startTime, endTime pgtype.Timestamptz
	if startStr != "" {
		t, err := time.Parse(time.RFC3339, startStr)
		if err == nil {
			startTime = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}
	if endStr != "" {
		t, err := time.Parse(time.RFC3339, endStr)
		if err == nil {
			endTime = pgtype.Timestamptz{Time: t, Valid: true}
		}
	}

	var statuses []string
	if statusesStr != "" {
		statuses = strings.Split(statusesStr, ",")
	}

	// List interviews
	params := repository.ListInterviewsParams{
		Limit:     limit,
		Offset:    offset,
		StartTime: startTime,
		EndTime:   endTime,
		Statuses:  statuses,
	}
	interviews, err := h.queries.ListInterviews(ctx, params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list interviews"})
		return
	}

	// Get total count
	countParams := repository.CountInterviewsParams{
		StartTime: startTime,
		EndTime:   endTime,
		Statuses:  statuses,
	}
	total, err := h.queries.CountInterviews(ctx, countParams)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count interviews"})
		return
	}

	result := make([]model.Interview, len(interviews))
	for i, interview := range interviews {
		var snapshot *model.SnapshotStatus
		if interview.SnapshotStatusKey != "" {
			snapshot = &model.SnapshotStatus{
				Key:   interview.SnapshotStatusKey,
				Label: interview.SnapshotStatusLabel,
			}
		}

		result[i] = model.Interview{
			ID:                 uuidToString(interview.ID),
			CandidateID:        uuidToString(interview.CandidateID),
			CandidateName:      interview.CandidateName,
			CandidateResumeURL: interview.CandidateResumeUrl,
			InterviewerID:      uuidToString(interview.InterviewerID),
			InterviewerName:    interview.InterviewerFirstName + " " + interview.InterviewerLastName,
			JobID:              uuidToString(interview.JobID),
			JobTitle:           interview.JobTitle,
			ScheduledTime:      interview.ScheduledTime.Time,
			ScheduledEndTime:   interview.ScheduledEndTime.Time,
			Status:             interview.Status,
			CreatedAt:          interview.CreatedAt.Time,
			SnapshotStatus:     snapshot,
		}
	}

	c.JSON(http.StatusOK, model.InterviewListResult{
		Interviews: result,
		Total:      total,
		Page:       page,
		Limit:      pageSize,
	})
}

// checkRecruiterOrAdmin checks if the user is an admin or a recruiter
func (h *RecruitmentHandler) checkRecruiterOrAdmin(ctx context.Context, userID pgtype.UUID) (bool, error) {
	res, err := h.queries.CheckRecruiterOrAdmin(ctx, userID)
	return res.Bool, err
}

// UpdateInterview updates interview details (Recruiter only)
func (h *RecruitmentHandler) UpdateInterview(c *gin.Context) {
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

	var req struct {
		InterviewerID    string    `json:"interviewerId" binding:"required"`
		ScheduledTime    time.Time `json:"scheduledTime" binding:"required"`
		ScheduledEndTime time.Time `json:"scheduledEndTime" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate time
	if validationErr := validateInterviewSchedule(req.ScheduledTime, req.ScheduledEndTime, time.Now()); validationErr != "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": validationErr})
		return
	}

	ctx := c.Request.Context()

	// Check permissions
	employee, err := h.queries.GetEmployeeByUserID(ctx, userID)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Employee profile not found"})
		return
	}

	recruiterID, err := h.queries.CheckRecruiterRole(ctx, employee.ID)
	isRecruiter := err == nil && recruiterID.Valid

	// Check Admin
	isAdmin, _ := h.queries.CheckIsAdmin(ctx, userID)

	if !isRecruiter && !isAdmin {
		c.JSON(http.StatusForbidden, gin.H{"error": "Permission denied"})
		return
	}

	interviewerUUID, err := parseUUID(req.InterviewerID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid interviewer ID"})
		return
	}

	updatedInterview, err := h.queries.UpdateInterview(ctx, repository.UpdateInterviewParams{
		ID:               interviewID,
		ScheduledTime:    pgtype.Timestamptz{Time: req.ScheduledTime, Valid: true},
		ScheduledEndTime: pgtype.Timestamptz{Time: req.ScheduledEndTime, Valid: true},
		InterviewerID:    interviewerUUID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update interview"})
		return
	}

	// Ensure the new interviewer has the role
	_ = h.queries.AssignInterviewerRole(ctx, interviewerUUID)

	var snapshot *model.SnapshotStatus
	if updatedInterview.SnapshotStatusKey != "" {
		snapshot = &model.SnapshotStatus{
			Key:   updatedInterview.SnapshotStatusKey,
			Label: updatedInterview.SnapshotStatusLabel,
		}
	}

	c.JSON(http.StatusOK, model.Interview{
		ID:               uuidToString(updatedInterview.ID),
		CandidateID:      uuidToString(updatedInterview.CandidateID),
		InterviewerID:    uuidToString(updatedInterview.InterviewerID),
		JobID:            uuidToString(updatedInterview.JobID),
		ScheduledTime:    updatedInterview.ScheduledTime.Time,
		ScheduledEndTime: updatedInterview.ScheduledEndTime.Time,
		Status:           updatedInterview.Status,
		CreatedAt:        updatedInterview.CreatedAt.Time,
		SnapshotStatus:   snapshot,
	})
}
