package mocks

import (
	"context"
	"hr-backend/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

type MockQuerier struct {
	CreateJobFunc             func(ctx context.Context, arg repository.CreateJobParams) (repository.Job, error)
	GetJobFunc                func(ctx context.Context, id pgtype.UUID) (repository.Job, error)
	ListJobsFunc              func(ctx context.Context) ([]repository.Job, error)
	UpdateJobFunc             func(ctx context.Context, arg repository.UpdateJobParams) (repository.Job, error)
	UpdateJobStatusFunc       func(ctx context.Context, arg repository.UpdateJobStatusParams) (repository.Job, error)
	DeleteJobFunc             func(ctx context.Context, id pgtype.UUID) error

	CreateCandidateFunc       func(ctx context.Context, arg repository.CreateCandidateParams) (repository.Candidate, error)
	GetCandidateFunc          func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error)
	ListCandidatesFunc        func(ctx context.Context, arg pgtype.UUID) ([]repository.ListCandidatesRow, error)
	UpdateCandidateFunc       func(ctx context.Context, arg repository.UpdateCandidateParams) (repository.Candidate, error)
	UpdateCandidateStatusFunc func(ctx context.Context, arg repository.UpdateCandidateStatusParams) (repository.Candidate, error)
	UpdateCandidateNoteFunc   func(ctx context.Context, arg repository.UpdateCandidateNoteParams) (repository.Candidate, error)
	UpdateCandidateResumeFunc func(ctx context.Context, arg repository.UpdateCandidateResumeParams) (repository.Candidate, error)
	DeleteCandidateFunc       func(ctx context.Context, id pgtype.UUID) error

	CreateUserFunc        func(ctx context.Context, arg repository.CreateUserParams) (repository.User, error)
	GetUserByUsernameFunc func(ctx context.Context, username string) (repository.User, error)
	GetUserByIDFunc       func(ctx context.Context, id pgtype.UUID) (repository.User, error)
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
func (m *MockQuerier) ListCandidates(ctx context.Context, arg pgtype.UUID) ([]repository.ListCandidatesRow, error) {
	return m.ListCandidatesFunc(ctx, arg)
}
func (m *MockQuerier) UpdateCandidate(ctx context.Context, arg repository.UpdateCandidateParams) (repository.Candidate, error) {
	return m.UpdateCandidateFunc(ctx, arg)
}
func (m *MockQuerier) UpdateCandidateStatus(ctx context.Context, arg repository.UpdateCandidateStatusParams) (repository.Candidate, error) {
	return m.UpdateCandidateStatusFunc(ctx, arg)
}
func (m *MockQuerier) UpdateCandidateNote(ctx context.Context, arg repository.UpdateCandidateNoteParams) (repository.Candidate, error) {
	return m.UpdateCandidateNoteFunc(ctx, arg)
}
func (m *MockQuerier) UpdateCandidateResume(ctx context.Context, arg repository.UpdateCandidateResumeParams) (repository.Candidate, error) {
	return m.UpdateCandidateResumeFunc(ctx, arg)
}
func (m *MockQuerier) DeleteCandidate(ctx context.Context, id pgtype.UUID) error {
	return m.DeleteCandidateFunc(ctx, id)
}

func (m *MockQuerier) CreateUser(ctx context.Context, arg repository.CreateUserParams) (repository.User, error) {
	return m.CreateUserFunc(ctx, arg)
}
func (m *MockQuerier) GetUserByUsername(ctx context.Context, username string) (repository.User, error) {
	return m.GetUserByUsernameFunc(ctx, username)
}
func (m *MockQuerier) GetUserByID(ctx context.Context, id pgtype.UUID) (repository.User, error) {
	return m.GetUserByIDFunc(ctx, id)
}