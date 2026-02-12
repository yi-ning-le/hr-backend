package mocks

import (
	"context"
	"hr-backend/internal/repository"

	"github.com/jackc/pgx/v5/pgtype"
)

type MockQuerier struct {
	CreateJobFunc       func(ctx context.Context, arg repository.CreateJobParams) (repository.Job, error)
	GetJobFunc          func(ctx context.Context, id pgtype.UUID) (repository.Job, error)
	ListJobsFunc        func(ctx context.Context) ([]repository.Job, error)
	UpdateJobFunc       func(ctx context.Context, arg repository.UpdateJobParams) (repository.Job, error)
	UpdateJobStatusFunc func(ctx context.Context, arg repository.UpdateJobStatusParams) (repository.Job, error)
	DeleteJobFunc       func(ctx context.Context, id pgtype.UUID) error

	AssignReviewerFunc        func(ctx context.Context, arg repository.AssignReviewerParams) (repository.AssignReviewerRow, error)
	SubmitReviewFunc          func(ctx context.Context, arg repository.SubmitReviewParams) (repository.SubmitReviewRow, error)
	CreateCandidateFunc       func(ctx context.Context, arg repository.CreateCandidateParams) (repository.Candidate, error)
	GetCandidateFunc          func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error)
	ListCandidatesFunc        func(ctx context.Context, arg repository.ListCandidatesParams) ([]repository.ListCandidatesRow, error)
	UpdateCandidateFunc       func(ctx context.Context, arg repository.UpdateCandidateParams) (repository.Candidate, error)
	UpdateCandidateStatusFunc func(ctx context.Context, arg repository.UpdateCandidateStatusParams) (repository.Candidate, error)
	UpdateCandidateResumeFunc func(ctx context.Context, arg repository.UpdateCandidateResumeParams) (repository.Candidate, error)
	DeleteCandidateFunc       func(ctx context.Context, id pgtype.UUID) error

	CountCandidatesFunc         func(ctx context.Context, arg repository.CountCandidatesParams) (int64, error)
	GetCandidateCountsByJobFunc func(ctx context.Context) ([]repository.GetCandidateCountsByJobRow, error)

	CreateUserFunc        func(ctx context.Context, arg repository.CreateUserParams) (repository.User, error)
	DeleteUserFunc        func(ctx context.Context, id pgtype.UUID) error
	GetUserByUsernameFunc func(ctx context.Context, username string) (repository.User, error)
	GetUserByIDFunc       func(ctx context.Context, id pgtype.UUID) (repository.User, error)

	// Employee mock functions
	CreateEmployeeFunc func(ctx context.Context, arg repository.CreateEmployeeParams) (repository.Employee, error)
	GetEmployeeFunc    func(ctx context.Context, id pgtype.UUID) (repository.Employee, error)
	ListEmployeesFunc  func(ctx context.Context, arg repository.ListEmployeesParams) ([]repository.Employee, error)
	CountEmployeesFunc func(ctx context.Context, arg repository.CountEmployeesParams) (int64, error)
	UpdateEmployeeFunc func(ctx context.Context, arg repository.UpdateEmployeeParams) (repository.Employee, error)
	DeleteEmployeeFunc func(ctx context.Context, id pgtype.UUID) error

	// Candidate Status mock functions
	ListCandidateStatusesFunc       func(ctx context.Context) ([]repository.CandidateStatus, error)
	GetCandidateStatusFunc          func(ctx context.Context, id pgtype.UUID) (repository.CandidateStatus, error)
	GetCandidateStatusBySlugFunc    func(ctx context.Context, slug string) (repository.CandidateStatus, error)
	CreateCandidateStatusFunc       func(ctx context.Context, arg repository.CreateCandidateStatusParams) (repository.CandidateStatus, error)
	UpdateCandidateStatusFieldsFunc func(ctx context.Context, arg repository.UpdateCandidateStatusFieldsParams) (repository.CandidateStatus, error)
	UpdateCandidateStatusOrderFunc  func(ctx context.Context, arg repository.UpdateCandidateStatusOrderParams) error
	DeleteCandidateStatusFunc       func(ctx context.Context, id pgtype.UUID) error

	// Recruitment mock functions
	AssignRecruiterRoleFunc         func(ctx context.Context, employeeID pgtype.UUID) error
	CheckIsAdminFunc                func(ctx context.Context, id pgtype.UUID) (bool, error)
	CheckRecruiterRoleFunc          func(ctx context.Context, employeeID pgtype.UUID) (pgtype.UUID, error)
	GetActiveInterviewCountFunc     func(ctx context.Context, interviewerID pgtype.UUID) (int64, error)
	GrantResumeReviewCapabilityFunc func(ctx context.Context, id pgtype.UUID) error
	RevokeRecruiterRoleFunc         func(ctx context.Context, employeeID pgtype.UUID) error
	ListRecruitersFunc              func(ctx context.Context) ([]repository.ListRecruitersRow, error)
	GetEmployeeByUserIDFunc         func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error)
	CreateInterviewFunc             func(ctx context.Context, arg repository.CreateInterviewParams) (repository.Interview, error)
	GetInterviewFunc                func(ctx context.Context, id pgtype.UUID) (repository.Interview, error)
	ListInterviewsByInterviewerFunc func(ctx context.Context, interviewerID pgtype.UUID) ([]repository.Interview, error)
	TransferInterviewFunc           func(ctx context.Context, arg repository.TransferInterviewParams) (repository.Interview, error)
	UpdateInterviewStatusFunc       func(ctx context.Context, arg repository.UpdateInterviewStatusParams) (repository.Interview, error)
	UpdateInterviewNoteFunc         func(ctx context.Context, arg repository.UpdateInterviewNoteParams) (repository.Interview, error)

	// HR role check
	CheckIsHRFunc    func(ctx context.Context, id pgtype.UUID) (bool, error)
	AssignHRRoleFunc func(ctx context.Context, id pgtype.UUID) error
	RevokeHRRoleFunc func(ctx context.Context, id pgtype.UUID) error
	ListHRsFunc      func(ctx context.Context) ([]repository.ListHRsRow, error)

	// Candidate Comment mock functions
	CreateCandidateCommentFunc func(ctx context.Context, arg repository.CreateCandidateCommentParams) (repository.CandidateComment, error)
	ListCandidateCommentsFunc  func(ctx context.Context, candidateID pgtype.UUID) ([]repository.ListCandidateCommentsRow, error)
	GetCandidateCommentFunc    func(ctx context.Context, id pgtype.UUID) (repository.CandidateComment, error)
	DeleteCandidateCommentFunc func(ctx context.Context, id pgtype.UUID) error
}

func (m *MockQuerier) CreateJob(ctx context.Context, arg repository.CreateJobParams) (repository.Job, error) {
	return m.CreateJobFunc(ctx, arg)
}
func (m *MockQuerier) GetJob(ctx context.Context, id pgtype.UUID) (repository.Job, error) {
	return m.GetJobFunc(ctx, id)
}
func (m *MockQuerier) ListJobs(ctx context.Context) ([]repository.Job, error) {
	return m.ListJobsFunc(ctx)
}
func (m *MockQuerier) UpdateJob(ctx context.Context, arg repository.UpdateJobParams) (repository.Job, error) {
	return m.UpdateJobFunc(ctx, arg)
}
func (m *MockQuerier) UpdateJobStatus(ctx context.Context, arg repository.UpdateJobStatusParams) (repository.Job, error) {
	return m.UpdateJobStatusFunc(ctx, arg)
}
func (m *MockQuerier) DeleteJob(ctx context.Context, id pgtype.UUID) error {
	return m.DeleteJobFunc(ctx, id)
}

func (m *MockQuerier) CreateCandidate(ctx context.Context, arg repository.CreateCandidateParams) (repository.Candidate, error) {
	return m.CreateCandidateFunc(ctx, arg)
}
func (m *MockQuerier) GetCandidate(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
	return m.GetCandidateFunc(ctx, id)
}
func (m *MockQuerier) ListCandidates(ctx context.Context, arg repository.ListCandidatesParams) ([]repository.ListCandidatesRow, error) {
	return m.ListCandidatesFunc(ctx, arg)
}
func (m *MockQuerier) UpdateCandidate(ctx context.Context, arg repository.UpdateCandidateParams) (repository.Candidate, error) {
	return m.UpdateCandidateFunc(ctx, arg)
}
func (m *MockQuerier) UpdateCandidateStatus(ctx context.Context, arg repository.UpdateCandidateStatusParams) (repository.Candidate, error) {
	return m.UpdateCandidateStatusFunc(ctx, arg)
}
func (m *MockQuerier) UpdateCandidateResume(ctx context.Context, arg repository.UpdateCandidateResumeParams) (repository.Candidate, error) {
	return m.UpdateCandidateResumeFunc(ctx, arg)
}
func (m *MockQuerier) DeleteCandidate(ctx context.Context, id pgtype.UUID) error {
	return m.DeleteCandidateFunc(ctx, id)
}

func (m *MockQuerier) CountCandidates(ctx context.Context, arg repository.CountCandidatesParams) (int64, error) {
	if m.CountCandidatesFunc != nil {
		return m.CountCandidatesFunc(ctx, arg)
	}
	return 0, nil
}

func (m *MockQuerier) GetCandidateCountsByJob(ctx context.Context) ([]repository.GetCandidateCountsByJobRow, error) {
	if m.GetCandidateCountsByJobFunc != nil {
		return m.GetCandidateCountsByJobFunc(ctx)
	}
	return nil, nil
}

func (m *MockQuerier) AssignReviewer(ctx context.Context, arg repository.AssignReviewerParams) (repository.AssignReviewerRow, error) {
	return m.AssignReviewerFunc(ctx, arg)
}

func (m *MockQuerier) SubmitReview(ctx context.Context, arg repository.SubmitReviewParams) (repository.SubmitReviewRow, error) {
	return m.SubmitReviewFunc(ctx, arg)
}

func (m *MockQuerier) CreateUser(ctx context.Context, arg repository.CreateUserParams) (repository.User, error) {
	return m.CreateUserFunc(ctx, arg)
}
func (m *MockQuerier) DeleteUser(ctx context.Context, id pgtype.UUID) error {
	if m.DeleteUserFunc != nil {
		return m.DeleteUserFunc(ctx, id)
	}
	return nil
}
func (m *MockQuerier) GetUserByUsername(ctx context.Context, username string) (repository.User, error) {
	return m.GetUserByUsernameFunc(ctx, username)
}
func (m *MockQuerier) GetUserByID(ctx context.Context, id pgtype.UUID) (repository.User, error) {
	return m.GetUserByIDFunc(ctx, id)
}

// Employee methods
func (m *MockQuerier) CreateEmployee(ctx context.Context, arg repository.CreateEmployeeParams) (repository.Employee, error) {
	if m.CreateEmployeeFunc != nil {
		return m.CreateEmployeeFunc(ctx, arg)
	}
	return repository.Employee{}, nil
}
func (m *MockQuerier) GetEmployee(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
	if m.GetEmployeeFunc != nil {
		return m.GetEmployeeFunc(ctx, id)
	}
	return repository.Employee{}, nil
}
func (m *MockQuerier) ListEmployees(ctx context.Context, arg repository.ListEmployeesParams) ([]repository.Employee, error) {
	if m.ListEmployeesFunc != nil {
		return m.ListEmployeesFunc(ctx, arg)
	}
	return nil, nil
}
func (m *MockQuerier) CountEmployees(ctx context.Context, arg repository.CountEmployeesParams) (int64, error) {
	if m.CountEmployeesFunc != nil {
		return m.CountEmployeesFunc(ctx, arg)
	}
	return 0, nil
}
func (m *MockQuerier) UpdateEmployee(ctx context.Context, arg repository.UpdateEmployeeParams) (repository.Employee, error) {
	if m.UpdateEmployeeFunc != nil {
		return m.UpdateEmployeeFunc(ctx, arg)
	}
	return repository.Employee{}, nil
}
func (m *MockQuerier) DeleteEmployee(ctx context.Context, id pgtype.UUID) error {
	if m.DeleteEmployeeFunc != nil {
		return m.DeleteEmployeeFunc(ctx, id)
	}
	return nil
}

// Candidate Status methods
func (m *MockQuerier) ListCandidateStatuses(ctx context.Context) ([]repository.CandidateStatus, error) {
	if m.ListCandidateStatusesFunc != nil {
		return m.ListCandidateStatusesFunc(ctx)
	}
	return nil, nil
}
func (m *MockQuerier) GetCandidateStatus(ctx context.Context, id pgtype.UUID) (repository.CandidateStatus, error) {
	if m.GetCandidateStatusFunc != nil {
		return m.GetCandidateStatusFunc(ctx, id)
	}
	return repository.CandidateStatus{}, nil
}
func (m *MockQuerier) GetCandidateStatusBySlug(ctx context.Context, slug string) (repository.CandidateStatus, error) {
	if m.GetCandidateStatusBySlugFunc != nil {
		return m.GetCandidateStatusBySlugFunc(ctx, slug)
	}
	return repository.CandidateStatus{}, nil
}
func (m *MockQuerier) CreateCandidateStatus(ctx context.Context, arg repository.CreateCandidateStatusParams) (repository.CandidateStatus, error) {
	if m.CreateCandidateStatusFunc != nil {
		return m.CreateCandidateStatusFunc(ctx, arg)
	}
	return repository.CandidateStatus{}, nil
}
func (m *MockQuerier) UpdateCandidateStatusFields(ctx context.Context, arg repository.UpdateCandidateStatusFieldsParams) (repository.CandidateStatus, error) {
	if m.UpdateCandidateStatusFieldsFunc != nil {
		return m.UpdateCandidateStatusFieldsFunc(ctx, arg)
	}
	return repository.CandidateStatus{}, nil
}
func (m *MockQuerier) UpdateCandidateStatusOrder(ctx context.Context, arg repository.UpdateCandidateStatusOrderParams) error {
	if m.UpdateCandidateStatusOrderFunc != nil {
		return m.UpdateCandidateStatusOrderFunc(ctx, arg)
	}
	return nil
}
func (m *MockQuerier) DeleteCandidateStatus(ctx context.Context, id pgtype.UUID) error {
	if m.DeleteCandidateStatusFunc != nil {
		return m.DeleteCandidateStatusFunc(ctx, id)
	}
	return nil
}

// Recruitment methods
func (m *MockQuerier) AssignRecruiterRole(ctx context.Context, employeeID pgtype.UUID) error {
	if m.AssignRecruiterRoleFunc != nil {
		return m.AssignRecruiterRoleFunc(ctx, employeeID)
	}
	return nil
}
func (m *MockQuerier) CheckIsAdmin(ctx context.Context, id pgtype.UUID) (bool, error) {
	if m.CheckIsAdminFunc != nil {
		return m.CheckIsAdminFunc(ctx, id)
	}
	return false, nil
}
func (m *MockQuerier) CheckRecruiterRole(ctx context.Context, employeeID pgtype.UUID) (pgtype.UUID, error) {
	if m.CheckRecruiterRoleFunc != nil {
		return m.CheckRecruiterRoleFunc(ctx, employeeID)
	}
	return pgtype.UUID{}, nil
}
func (m *MockQuerier) GetActiveInterviewCount(ctx context.Context, interviewerID pgtype.UUID) (int64, error) {
	if m.GetActiveInterviewCountFunc != nil {
		return m.GetActiveInterviewCountFunc(ctx, interviewerID)
	}
	return 0, nil
}
func (m *MockQuerier) GrantResumeReviewCapability(ctx context.Context, id pgtype.UUID) error {
	if m.GrantResumeReviewCapabilityFunc != nil {
		return m.GrantResumeReviewCapabilityFunc(ctx, id)
	}
	return nil
}
func (m *MockQuerier) RevokeRecruiterRole(ctx context.Context, employeeID pgtype.UUID) error {
	if m.RevokeRecruiterRoleFunc != nil {
		return m.RevokeRecruiterRoleFunc(ctx, employeeID)
	}
	return nil
}
func (m *MockQuerier) ListRecruiters(ctx context.Context) ([]repository.ListRecruitersRow, error) {
	if m.ListRecruitersFunc != nil {
		return m.ListRecruitersFunc(ctx)
	}
	return nil, nil
}
func (m *MockQuerier) GetEmployeeByUserID(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
	if m.GetEmployeeByUserIDFunc != nil {
		return m.GetEmployeeByUserIDFunc(ctx, userID)
	}
	return repository.Employee{}, nil
}
func (m *MockQuerier) CreateInterview(ctx context.Context, arg repository.CreateInterviewParams) (repository.Interview, error) {
	if m.CreateInterviewFunc != nil {
		return m.CreateInterviewFunc(ctx, arg)
	}
	return repository.Interview{}, nil
}
func (m *MockQuerier) GetInterview(ctx context.Context, id pgtype.UUID) (repository.Interview, error) {
	if m.GetInterviewFunc != nil {
		return m.GetInterviewFunc(ctx, id)
	}
	return repository.Interview{}, nil
}
func (m *MockQuerier) ListInterviewsByInterviewer(ctx context.Context, interviewerID pgtype.UUID) ([]repository.Interview, error) {
	if m.ListInterviewsByInterviewerFunc != nil {
		return m.ListInterviewsByInterviewerFunc(ctx, interviewerID)
	}
	return nil, nil
}
func (m *MockQuerier) TransferInterview(ctx context.Context, arg repository.TransferInterviewParams) (repository.Interview, error) {
	if m.TransferInterviewFunc != nil {
		return m.TransferInterviewFunc(ctx, arg)
	}
	return repository.Interview{}, nil
}
func (m *MockQuerier) UpdateInterviewStatus(ctx context.Context, arg repository.UpdateInterviewStatusParams) (repository.Interview, error) {
	if m.UpdateInterviewStatusFunc != nil {
		return m.UpdateInterviewStatusFunc(ctx, arg)
	}
	return repository.Interview{}, nil
}
func (m *MockQuerier) UpdateInterviewNote(ctx context.Context, arg repository.UpdateInterviewNoteParams) (repository.Interview, error) {
	if m.UpdateInterviewNoteFunc != nil {
		return m.UpdateInterviewNoteFunc(ctx, arg)
	}
	return repository.Interview{}, nil
}

// CheckIsHR method for HR role check
func (m *MockQuerier) CheckIsHR(ctx context.Context, id pgtype.UUID) (bool, error) {
	if m.CheckIsHRFunc != nil {
		return m.CheckIsHRFunc(ctx, id)
	}
	return false, nil
}

// AssignHRRole assigns HR role to an employee
func (m *MockQuerier) AssignHRRole(ctx context.Context, id pgtype.UUID) error {
	if m.AssignHRRoleFunc != nil {
		return m.AssignHRRoleFunc(ctx, id)
	}
	return nil
}

// RevokeHRRole revokes HR role from an employee
func (m *MockQuerier) RevokeHRRole(ctx context.Context, id pgtype.UUID) error {
	if m.RevokeHRRoleFunc != nil {
		return m.RevokeHRRoleFunc(ctx, id)
	}
	return nil
}

// ListHRs returns all HR employees
func (m *MockQuerier) ListHRs(ctx context.Context) ([]repository.ListHRsRow, error) {
	if m.ListHRsFunc != nil {
		return m.ListHRsFunc(ctx)
	}
	return nil, nil
}

// Candidate Comment methods
func (m *MockQuerier) CreateCandidateComment(ctx context.Context, arg repository.CreateCandidateCommentParams) (repository.CandidateComment, error) {
	if m.CreateCandidateCommentFunc != nil {
		return m.CreateCandidateCommentFunc(ctx, arg)
	}
	return repository.CandidateComment{}, nil
}

func (m *MockQuerier) ListCandidateComments(ctx context.Context, candidateID pgtype.UUID) ([]repository.ListCandidateCommentsRow, error) {
	if m.ListCandidateCommentsFunc != nil {
		return m.ListCandidateCommentsFunc(ctx, candidateID)
	}
	return nil, nil
}

func (m *MockQuerier) GetCandidateComment(ctx context.Context, id pgtype.UUID) (repository.CandidateComment, error) {
	if m.GetCandidateCommentFunc != nil {
		return m.GetCandidateCommentFunc(ctx, id)
	}
	return repository.CandidateComment{}, nil
}

func (m *MockQuerier) DeleteCandidateComment(ctx context.Context, id pgtype.UUID) error {
	if m.DeleteCandidateCommentFunc != nil {
		return m.DeleteCandidateCommentFunc(ctx, id)
	}
	return nil
}
