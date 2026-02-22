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
	AppliedAt       time.Time `json:"appliedAt"`
	ReviewerID      string    `json:"reviewerId,omitempty"`
	ReviewStatus    string    `json:"reviewStatus,omitempty"`
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
	AppliedAt       time.Time `json:"appliedAt" binding:"required"`
	ReviewerID      string    `json:"reviewerId"`
}

// --- Auth Models ---

type LoginInput struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type RegisterInput struct {
	Username  string `json:"username" binding:"required,min=3,max=20,alphanum"`
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6,max=50"`
	FirstName string `json:"firstName" binding:"required"`
	LastName  string `json:"lastName" binding:"required"`
	Phone     string `json:"phone" binding:"required"`
}

type User struct {
	ID        string    `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	Avatar    string    `json:"avatar,omitempty"`
	CreatedAt time.Time `json:"createdAt"`
}

type AuthResponse struct {
	Token     string `json:"token"`
	SessionID string `json:"sessionId"`
	User      User   `json:"user"`
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
	// Only returned by employee creation API so HR can deliver first-login credentials.
	TemporaryPassword string `json:"temporaryPassword,omitempty"`
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

type Interview struct {
	ID                 string          `json:"id"`
	CandidateID        string          `json:"candidateId"`
	CandidateName      string          `json:"candidateName,omitempty"`
	CandidateResumeURL string          `json:"candidateResumeUrl,omitempty"`
	InterviewerID      string          `json:"interviewerId"`
	InterviewerName    string          `json:"interviewerName,omitempty"`
	JobID              string          `json:"jobId"`
	JobTitle           string          `json:"jobTitle,omitempty"`
	ScheduledTime      time.Time       `json:"scheduledTime"`
	ScheduledEndTime   time.Time       `json:"scheduledEndTime"`
	Status             string          `json:"status"`
	CreatedAt          time.Time       `json:"createdAt"`
	SnapshotStatus     *SnapshotStatus `json:"snapshotStatus,omitempty"`
}

type InterviewListResult struct {
	Interviews []Interview `json:"interviews"`
	Total      int64       `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
}

type SnapshotStatus struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

type CreateInterviewInput struct {
	CandidateID      string    `json:"candidateId" binding:"required"`
	InterviewerID    string    `json:"interviewerId" binding:"required"`
	JobID            string    `json:"jobId" binding:"required"`
	ScheduledTime    time.Time `json:"scheduledTime" binding:"required"`
	ScheduledEndTime time.Time `json:"scheduledEndTime" binding:"required"`
}

// --- Candidate Comment Models ---

type CandidateComment struct {
	ID           string    `json:"id"`
	CandidateID  string    `json:"candidateId"`
	AuthorID     string    `json:"authorId"`
	AuthorName   string    `json:"authorName"`
	AuthorAvatar string    `json:"authorAvatar,omitempty"`
	AuthorRole   string    `json:"authorRole"` // HR | INTERVIEWER
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"createdAt"`
}

type CreateCommentInput struct {
	Content string `json:"content" binding:"required"`
}

type SessionInfo struct {
	ID         string      `json:"id"`
	UserID     string      `json:"userId"`
	DeviceInfo interface{} `json:"deviceInfo"`
	IPAddress  string      `json:"ipAddress"`
	UserAgent  string      `json:"userAgent"`
	CreatedAt  time.Time   `json:"createdAt"`
	ExpiresAt  time.Time   `json:"expiresAt,omitempty"`
	IsActive   bool        `json:"isActive"`
}

type SessionListResponse struct {
	Sessions []SessionInfo `json:"sessions"`
}
