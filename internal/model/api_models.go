package model

import "time"

// JobPosition matches the JobPosition schema
type JobPosition struct {
	ID             string    `json:"id"`
	Title          string    `json:"title"`
	Department     string    `json:"department"`
	HeadCount      int       `json:"headCount"`
	OpenDate       time.Time `json:"openDate"`
	JobDescription string    `json:"jobDescription"`
	Note           string    `json:"note,omitempty"`
	Status         string    `json:"status"`
}

// JobInput matches the JobInput schema
type JobInput struct {
	Title          string    `json:"title" binding:"required"`
	Department     string    `json:"department" binding:"required"`
	HeadCount      int       `json:"headCount" binding:"required"`
	OpenDate       time.Time `json:"openDate" binding:"required"`
	JobDescription string    `json:"jobDescription" binding:"required"`
	Note           string    `json:"note"`
	Status         string    `json:"status"` // Default OPEN handled in service if missing
}

// Candidate matches the Candidate schema
type Candidate struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Avatar          string    `json:"avatar,omitempty"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	ExperienceYears int       `json:"experienceYears"`
	Education       string    `json:"education"`
	AppliedJobID    string    `json:"appliedJobId"`
	AppliedJobTitle string    `json:"appliedJobTitle"`
	Channel         string    `json:"channel"`
	ResumeURL       string    `json:"resumeUrl"`
	Status          string    `json:"status"`
	Note            string    `json:"note,omitempty"`
	AppliedAt       time.Time `json:"appliedAt"`
}

// CandidateInput matches the CandidateInput schema
type CandidateInput struct {
	Name            string    `json:"name" binding:"required"`
	Avatar          string    `json:"avatar"`
	Email           string    `json:"email" binding:"required,email"`
	Phone           string    `json:"phone" binding:"required"`
	ExperienceYears int       `json:"experienceYears" binding:"required"`
	Education       string    `json:"education" binding:"required"`
	AppliedJobID    string    `json:"appliedJobId" binding:"required"`
	AppliedJobTitle string    `json:"appliedJobTitle"` // Optional in input, but useful
	Channel         string    `json:"channel" binding:"required"`
	ResumeURL       string    `json:"resumeUrl" binding:"required"`
	Status          string    `json:"status"` // Default new
	Note            string    `json:"note"`
	AppliedAt       time.Time `json:"appliedAt" binding:"required"`
}

// --- Auth Models ---

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterInput struct {
	Username string `json:"username" binding:"required,min=3,max=20,alphanum"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6,max=50"`
}

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

// --- Employee Models ---

type Employee struct {
	ID             string    `json:"id"`
	FirstName      string    `json:"firstName"`
	LastName       string    `json:"lastName"`
	Email          string    `json:"email"`
	Phone          string    `json:"phone"`
	Department     string    `json:"department"`
	Position       string    `json:"position"`
	Status         string    `json:"status"`
	EmploymentType string    `json:"employmentType"`
	EmployeeType   string    `json:"employeeType"` // HR, EMPLOYEE
	JoinDate       time.Time `json:"joinDate"`
	ManagerID      string    `json:"managerId,omitempty"`
	UserID         string    `json:"userId,omitempty"`
}

type EmployeeInput struct {
	FirstName      string    `json:"firstName" binding:"required"`
	LastName       string    `json:"lastName" binding:"required"`
	Email          string    `json:"email" binding:"required,email"`
	Phone          string    `json:"phone" binding:"required"`
	Department     string    `json:"department" binding:"required"`
	Position       string    `json:"position" binding:"required"`
	Status         string    `json:"status"`
	EmploymentType string    `json:"employmentType"`
	EmployeeType   string    `json:"employeeType"` // HR, EMPLOYEE (default)
	JoinDate       time.Time `json:"joinDate" binding:"required"`
	ManagerID      string    `json:"managerId"`
	UserID         string    `json:"userId"`
}

type EmployeeListResult struct {
	Employees []Employee `json:"employees"`
	Total     int64      `json:"total"`
	Page      int        `json:"page"`
	Limit     int        `json:"limit"`
}

// --- Interview Models ---

type Interview struct {
	ID            string    `json:"id"`
	CandidateID   string    `json:"candidateId"`
	InterviewerID string    `json:"interviewerId"`
	JobID         string    `json:"jobId"`
	ScheduledTime time.Time `json:"scheduledTime"`
	Status        string    `json:"status"`
	Notes         string    `json:"notes,omitempty"`
	CreatedAt     time.Time `json:"createdAt"`
}

type CreateInterviewInput struct {
	CandidateID   string    `json:"candidateId" binding:"required"`
	InterviewerID string    `json:"interviewerId" binding:"required"`
	JobID         string    `json:"jobId" binding:"required"` // Can be derived from candidate, but explicit is okay
	ScheduledTime time.Time `json:"scheduledTime" binding:"required"`
	Notes         string    `json:"notes"`
}

type UpdateInterviewNotesInput struct {
	Notes string `json:"notes" binding:"required"`
}
