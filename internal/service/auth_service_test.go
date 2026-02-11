package service

import (
	"context"
	"testing"
	"time"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/test/mocks"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestAuthService_Register(t *testing.T) {
	mockRepo := &mocks.MockQuerier{}
	service := NewAuthService(mockRepo, "secret")

	ctx := context.Background()
	input := model.RegisterInput{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "password123",
	}

	// Mock CreateUser
	mockRepo.CreateUserFunc = func(ctx context.Context, arg repository.CreateUserParams) (repository.User, error) {
		assert.Equal(t, input.Username, arg.Username)
		assert.Equal(t, input.Email, arg.Email)
		return repository.User{
			ID:        pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			Username:  input.Username,
			Email:     input.Email,
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil
	}

	// Mock CreateEmployee - Verify it is called with correct params
	employeeCreated := false
	mockRepo.CreateEmployeeFunc = func(ctx context.Context, arg repository.CreateEmployeeParams) (repository.Employee, error) {
		employeeCreated = true
		assert.Equal(t, input.Username, arg.FirstName)
		assert.Equal(t, "Unassigned", arg.Department)
		assert.True(t, arg.UserID.Valid)
		assert.Equal(t, [16]byte{1}, arg.UserID.Bytes)

		return repository.Employee{
			ID:        pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
			UserID:    arg.UserID,
			FirstName: arg.FirstName,
		}, nil
	}

	// Mock GetUserByUsername (if needed, but Register doesn't call it in current impl)
	// If Register logic changes to check duplicates, adding this might be needed.

	user, err := service.Register(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, input.Username, user.Username)
	assert.True(t, employeeCreated, "CreateEmployee should have been called")
}

func TestAuthService_Register_RollsBackUserWhenEmployeeCreateFails(t *testing.T) {
	mockRepo := &mocks.MockQuerier{}
	service := NewAuthService(mockRepo, "secret")

	input := model.RegisterInput{
		Username: "rollback-user",
		Email:    "rollback@example.com",
		Password: "password123",
	}

	createdUserID := pgtype.UUID{Bytes: [16]byte{9}, Valid: true}
	mockRepo.CreateUserFunc = func(ctx context.Context, arg repository.CreateUserParams) (repository.User, error) {
		return repository.User{
			ID:        createdUserID,
			Username:  input.Username,
			Email:     input.Email,
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil
	}

	mockRepo.CreateEmployeeFunc = func(ctx context.Context, arg repository.CreateEmployeeParams) (repository.Employee, error) {
		return repository.Employee{}, assert.AnError
	}

	deleteCalled := false
	mockRepo.DeleteUserFunc = func(ctx context.Context, id pgtype.UUID) error {
		deleteCalled = true
		assert.Equal(t, createdUserID, id)
		return nil
	}

	user, err := service.Register(context.Background(), input)

	assert.Error(t, err)
	assert.Nil(t, user)
	assert.True(t, deleteCalled, "DeleteUser should be called for rollback")
}
