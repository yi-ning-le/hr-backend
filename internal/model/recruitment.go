package model

// RecruitmentRoleResponse is returned by GET /recruitment/role
type RecruitmentRoleResponse struct {
	IsAdmin       bool `json:"isAdmin"`
	IsRecruiter   bool `json:"isRecruiter"`
	IsInterviewer bool `json:"isInterviewer"`
}

// Recruiter represents an employee with recruiter role
type Recruiter struct {
	EmployeeID string `json:"employeeId"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Department string `json:"department"`
	Avatar     string `json:"avatar,omitempty"`
}

// TransferInterviewInput is the request body for transferring an interview
type TransferInterviewInput struct {
	NewInterviewerID string `json:"newInterviewerId" binding:"required"`
}
