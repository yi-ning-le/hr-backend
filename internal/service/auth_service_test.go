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
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "password123",
		FirstName: "Test",
		LastName:  "User",
		Phone:     "1234567890",
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
		assert.Equal(t, input.FirstName, arg.FirstName)
		assert.Equal(t, input.LastName, arg.LastName)
		assert.Equal(t, input.Phone, arg.Phone)
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
		Username:  "rollback-user",
		Email:     "rollback@example.com",
		Password:  "password123",
		FirstName: "Rollback",
		LastName:  "User",
		Phone:     "0000000000",
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

func TestAuthService_CleanupExpiredSessions(t *testing.T) {
	mockRepo := &mocks.MockQuerier{}
	service := NewAuthService(mockRepo, "secret")

	expiredCleanupCalled := false
	mockRepo.DeleteExpiredSessionsFunc = func(ctx context.Context) error {
		expiredCleanupCalled = true
		return nil
	}
	inactiveCleanupCalled := false
	mockRepo.DeleteInactiveSessionsFunc = func(ctx context.Context, lastActiveAt pgtype.Timestamptz) error {
		inactiveCleanupCalled = true
		assert.True(t, lastActiveAt.Valid)
		return nil
	}

	err := service.CleanupExpiredSessions(context.Background())

	assert.NoError(t, err)
	assert.True(t, expiredCleanupCalled, "DeleteExpiredSessions should have been called")
	assert.True(t, inactiveCleanupCalled, "DeleteInactiveSessions should have been called")
}

func TestAuthService_RefreshToken_RejectsMismatchedUserAgent(t *testing.T) {
	mockRepo := &mocks.MockQuerier{}
	service := NewAuthService(mockRepo, "secret")

	sessionID := "11111111-1111-1111-1111-111111111111"
	sessionUUID := pgtype.UUID{
		Bytes: [16]byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
		Valid: true,
	}
	userUUID := pgtype.UUID{
		Bytes: [16]byte{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22},
		Valid: true,
	}

	mockRepo.RefreshTokenFunc = func(ctx context.Context, id pgtype.UUID) (repository.Session, error) {
		assert.Equal(t, sessionUUID, id)
		return repository.Session{
			ID:        sessionUUID,
			UserID:    userUUID,
			UserAgent: pgtype.Text{String: "expected-agent", Valid: true},
			IsActive:  pgtype.Bool{Bool: true, Valid: true},
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(1 * time.Hour), Valid: true},
		}, nil
	}

	updateCalled := false
	mockRepo.DeleteSessionFunc = func(ctx context.Context, id pgtype.UUID) error {
		updateCalled = true
		return nil
	}
	mockRepo.GetUserByIDFunc = func(ctx context.Context, id pgtype.UUID) (repository.User, error) {
		t.Fatalf("GetUserByID should not be called on context mismatch")
		return repository.User{}, nil
	}

	resp, err := service.RefreshToken(context.Background(), sessionID, "different-agent")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "session context mismatch", err.Error())
	assert.False(t, updateCalled, "DeleteSession should not run on context mismatch")
}

func TestAuthService_RefreshToken_RejectsMissingUserAgent(t *testing.T) {
	mockRepo := &mocks.MockQuerier{}
	service := NewAuthService(mockRepo, "secret")

	sessionID := "11111111-1111-1111-1111-111111111111"
	sessionUUID := pgtype.UUID{
		Bytes: [16]byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
		Valid: true,
	}
	userUUID := pgtype.UUID{
		Bytes: [16]byte{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22},
		Valid: true,
	}

	mockRepo.RefreshTokenFunc = func(ctx context.Context, id pgtype.UUID) (repository.Session, error) {
		assert.Equal(t, sessionUUID, id)
		return repository.Session{
			ID:        sessionUUID,
			UserID:    userUUID,
			UserAgent: pgtype.Text{String: "bound-agent", Valid: true},
			IsActive:  pgtype.Bool{Bool: true, Valid: true},
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(1 * time.Hour), Valid: true},
		}, nil
	}

	updateCalled := false
	mockRepo.DeleteSessionFunc = func(ctx context.Context, id pgtype.UUID) error {
		updateCalled = true
		return nil
	}
	mockRepo.GetUserByIDFunc = func(ctx context.Context, id pgtype.UUID) (repository.User, error) {
		t.Fatalf("GetUserByID should not be called when user agent is missing")
		return repository.User{}, nil
	}

	resp, err := service.RefreshToken(context.Background(), sessionID, "")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, "session context mismatch", err.Error())
	assert.False(t, updateCalled, "DeleteSession should not run when user agent is missing")
}

func TestAuthService_RefreshToken_Success(t *testing.T) {
	mockRepo := &mocks.MockQuerier{}
	service := NewAuthService(mockRepo, "secret")

	oldSessionID := "11111111-1111-1111-1111-111111111111"
	oldSessionUUID := pgtype.UUID{
		Bytes: [16]byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
		Valid: true,
	}
	userUUID := pgtype.UUID{
		Bytes: [16]byte{0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33},
		Valid: true,
	}
	newSessionUUID := pgtype.UUID{
		Bytes: [16]byte{0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99, 0x99},
		Valid: true,
	}

	mockRepo.RefreshTokenFunc = func(ctx context.Context, id pgtype.UUID) (repository.Session, error) {
		assert.Equal(t, oldSessionUUID, id)
		return repository.Session{
			ID:        oldSessionUUID,
			UserID:    userUUID,
			UserAgent: pgtype.Text{String: "same-agent", Valid: true},
			IsActive:  pgtype.Bool{Bool: true, Valid: true},
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(1 * time.Hour), Valid: true},
		}, nil
	}

	deleteCalled := false
	mockRepo.DeleteSessionFunc = func(ctx context.Context, id pgtype.UUID) error {
		deleteCalled = true
		assert.Equal(t, oldSessionUUID, id)
		return nil
	}

	createCalled := false
	mockRepo.CreateSessionFunc = func(ctx context.Context, arg repository.CreateSessionParams) (repository.Session, error) {
		createCalled = true
		assert.Equal(t, userUUID, arg.UserID)
		return repository.Session{
			ID:     newSessionUUID,
			UserID: userUUID,
		}, nil
	}

	mockRepo.GetUserByIDFunc = func(ctx context.Context, id pgtype.UUID) (repository.User, error) {
		assert.Equal(t, userUUID, id)
		return repository.User{
			ID:        userUUID,
			Username:  "test-user",
			Email:     "test@example.com",
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
			Avatar:    pgtype.Text{Valid: false},
		}, nil
	}

	resp, err := service.RefreshToken(context.Background(), oldSessionID, "same-agent")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "99999999-9999-9999-9999-999999999999", resp.SessionID)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, "test-user", resp.User.Username)
	assert.True(t, deleteCalled, "DeleteSession should be called")
	assert.True(t, createCalled, "CreateSession should be called")
}

func TestAuthService_RefreshToken_AllowsLegacySessionWithoutUserAgent(t *testing.T) {
	mockRepo := &mocks.MockQuerier{}
	service := NewAuthService(mockRepo, "secret")

	oldSessionID := "11111111-1111-1111-1111-111111111111"
	oldSessionUUID := pgtype.UUID{
		Bytes: [16]byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11},
		Valid: true,
	}
	userUUID := pgtype.UUID{
		Bytes: [16]byte{0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44},
		Valid: true,
	}
	newSessionUUID := pgtype.UUID{
		Bytes: [16]byte{0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55},
		Valid: true,
	}

	mockRepo.RefreshTokenFunc = func(ctx context.Context, id pgtype.UUID) (repository.Session, error) {
		assert.Equal(t, oldSessionUUID, id)
		return repository.Session{
			ID:        oldSessionUUID,
			UserID:    userUUID,
			UserAgent: pgtype.Text{Valid: false},
			IsActive:  pgtype.Bool{Bool: true, Valid: true},
			ExpiresAt: pgtype.Timestamptz{Time: time.Now().Add(1 * time.Hour), Valid: true},
		}, nil
	}
	mockRepo.DeleteSessionFunc = func(ctx context.Context, id pgtype.UUID) error {
		assert.Equal(t, oldSessionUUID, id)
		return nil
	}
	mockRepo.CreateSessionFunc = func(ctx context.Context, arg repository.CreateSessionParams) (repository.Session, error) {
		assert.Equal(t, userUUID, arg.UserID)
		assert.True(t, arg.UserAgent.Valid)
		assert.Equal(t, "rotated-agent", arg.UserAgent.String)
		return repository.Session{
			ID:     newSessionUUID,
			UserID: userUUID,
		}, nil
	}
	mockRepo.GetUserByIDFunc = func(ctx context.Context, id pgtype.UUID) (repository.User, error) {
		assert.Equal(t, userUUID, id)
		return repository.User{
			ID:        userUUID,
			Username:  "legacy-user",
			Email:     "legacy@example.com",
			CreatedAt: pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil
	}

	resp, err := service.RefreshToken(context.Background(), oldSessionID, "rotated-agent")

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "55555555-5555-5555-5555-555555555555", resp.SessionID)
}
