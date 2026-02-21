package model

// RecruitmentRoleResponse is returned by GET /recruitment/role
type RecruitmentRoleResponse struct {
	IsAdmin       bool `json:"isAdmin"`
	IsRecruiter   bool `json:"isRecruiter"`
	IsInterviewer bool `json:"isInterviewer"`
	IsHR          bool `json:"isHR"`
}

// Recruiter represents an employee with recruiter role
type Recruiter struct {
	EmployeeID string `json:"employeeId"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Department string `json:"department"`
	Avatar     string `json:"avatar,omitempty"`
}

// Interviewer represents an employee with interviewer role
type Interviewer struct {
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

// UpdateInterviewStatusInput is the request body for updating interview status
type UpdateInterviewStatusInput struct {
	Status string `json:"status" binding:"required,oneof=COMPLETED CANCELLED"`
}

// HREmployee represents an employee with HR role
type HREmployee struct {
	EmployeeID string `json:"employeeId"`
	FirstName  string `json:"firstName"`
	LastName   string `json:"lastName"`
	Department string `json:"department"`
}
