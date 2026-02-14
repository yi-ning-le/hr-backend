package service

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"time"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

const SessionDuration = 24 * time.Hour

type DeviceInfo struct {
	Browser   string `json:"browser"`
	OS        string `json:"os"`
	Platform  string `json:"platform"`
	IP        string `json:"ip"`
	UserAgent string `json:"userAgent"`
}

type AuthService struct {
	repo       repository.Querier
	txBeginner TxBeginner
	jwtSecret  string
}

func NewAuthService(repo repository.Querier, jwtSecret string, txBeginner ...TxBeginner) *AuthService {
	var beginner TxBeginner
	if len(txBeginner) > 0 {
		beginner = txBeginner[0]
	}

	return &AuthService{
		repo:       repo,
		txBeginner: beginner,
		jwtSecret:  jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, input model.RegisterInput) (*model.User, error) {
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	params := repository.CreateUserParams{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: hashedPassword,
		Avatar:       pgtype.Text{Valid: false},
	}

	firstName := input.Username
	lastName := "User"
	now := pgtype.Timestamptz{
		Time:  utils.Now(),
		Valid: true,
	}

	empParams := repository.CreateEmployeeParams{
		FirstName:      firstName,
		LastName:       lastName,
		Email:          input.Email,
		Phone:          "",
		Department:     "Unassigned",
		Position:       "New Hire",
		Status:         "Active",
		EmploymentType: "FullTime",
		JoinDate:       now,
		UserID:         pgtype.UUID{},
		ManagerID:      pgtype.UUID{Valid: false},
	}

	var user repository.User
	if s.txBeginner != nil {
		err = runInTx(ctx, s.txBeginner, func(txQueries *repository.Queries) error {
			createdUser, createErr := txQueries.CreateUser(ctx, params)
			if createErr != nil {
				return createErr
			}
			user = createdUser
			empParams.UserID = createdUser.ID

			if _, createEmpErr := txQueries.CreateEmployee(ctx, empParams); createEmpErr != nil {
				return errors.New("failed to create linked employee record: " + createEmpErr.Error())
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
	} else {
		createdUser, createErr := s.repo.CreateUser(ctx, params)
		if createErr != nil {
			return nil, createErr
		}
		user = createdUser
		empParams.UserID = createdUser.ID

		if _, createEmpErr := s.repo.CreateEmployee(ctx, empParams); createEmpErr != nil {
			_ = s.repo.DeleteUser(ctx, createdUser.ID)
			return nil, errors.New("failed to create linked employee record: " + createEmpErr.Error())
		}
	}

	return &model.User{
		ID:        utils.UUIDToString(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, input model.LoginInput) (*model.AuthResponse, error) {
	return s.LoginWithDevice(ctx, input, DeviceInfo{})
}

func (s *AuthService) LoginWithDevice(ctx context.Context, input model.LoginInput, device DeviceInfo) (*model.AuthResponse, error) {
	user, err := s.repo.GetUserByUsername(ctx, input.Username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	userID := user.ID
	deviceJSON, _ := json.Marshal(device)
	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(SessionDuration),
		Valid: true,
	}

	session, err := s.repo.CreateSession(ctx, repository.CreateSessionParams{
		UserID:     userID,
		DeviceInfo: deviceJSON,
		IpAddress:  pgtype.Text{String: device.IP, Valid: device.IP != ""},
		UserAgent:  pgtype.Text{String: device.UserAgent, Valid: device.UserAgent != ""},
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		return nil, errors.New("failed to create session: " + err.Error())
	}

	token, err := utils.GenerateTokenWithSession(
		utils.UUIDToString(user.ID),
		user.Username,
		utils.UUIDToString(session.ID),
		s.jwtSecret,
	)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token:     token,
		SessionID: utils.UUIDToString(session.ID),
		User: model.User{
			ID:        utils.UUIDToString(user.ID),
			Username:  user.Username,
			Email:     user.Email,
			Avatar:    user.Avatar.String,
			CreatedAt: user.CreatedAt.Time,
		},
	}, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID string) error {
	sessionUUID, err := utils.StringToUUID(sessionID)
	if err != nil {
		return errors.New("invalid session ID")
	}

	_, err = s.repo.GetActiveSessionByID(ctx, sessionUUID)
	if err != nil {
		return errors.New("session not found or already expired")
	}

	return s.repo.DeactivateSession(ctx, sessionUUID)
}

func (s *AuthService) LogoutAll(ctx context.Context, userID string) error {
	userUUID, err := utils.StringToUUID(userID)
	if err != nil {
		return errors.New("invalid user ID")
	}

	return s.repo.DeactivateUserSessions(ctx, userUUID)
}

func (s *AuthService) GetSessions(ctx context.Context, userID string) ([]model.SessionInfo, error) {
	userUUID, err := utils.StringToUUID(userID)
	if err != nil {
		return nil, errors.New("invalid user ID")
	}

	sessions, err := s.repo.GetUserSessions(ctx, userUUID)
	if err != nil {
		return nil, err
	}

	var result []model.SessionInfo
	for _, session := range sessions {
		var deviceInfo DeviceInfo
		_ = json.Unmarshal(session.DeviceInfo, &deviceInfo)

		result = append(result, model.SessionInfo{
			ID:         utils.UUIDToString(session.ID),
			UserID:     utils.UUIDToString(session.UserID),
			DeviceInfo: deviceInfo,
			IPAddress:  session.IpAddress.String,
			UserAgent:  session.UserAgent.String,
			CreatedAt:  session.CreatedAt.Time,
			ExpiresAt:  session.ExpiresAt.Time,
			IsActive:   session.IsActive.Bool,
		})
	}

	return result, nil
}

func (s *AuthService) ValidateSession(ctx context.Context, sessionID string) error {
	sessionUUID, err := utils.StringToUUID(sessionID)
	if err != nil {
		return errors.New("invalid session ID")
	}

	_, err = s.repo.GetActiveSessionByID(ctx, sessionUUID)
	if err != nil {
		return errors.New("session not found or expired")
	}

	return nil
}

func (s *AuthService) CleanupExpiredSessions(ctx context.Context) error {
	// 1. Delete expired sessions (expires_at < now)
	if err := s.repo.DeleteExpiredSessions(ctx); err != nil {
		return err
	}

	// 2. Delete inactive sessions (last_active_at < now - 7 days)
	// Calculate the timestamp for 7 days ago
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)

	// Create a pgtype.Timestamptz from the calculated time
	pgTimestamp := pgtype.Timestamptz{
		Time:  sevenDaysAgo,
		Valid: true,
	}

	// Call the repository method with the correctly typed timestamp
	return s.repo.DeleteInactiveSessions(ctx, pgTimestamp)
}

func (s *AuthService) StartCleanupTask(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	log.Printf("Start Cleanup task")

	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := s.CleanupExpiredSessions(ctx); err != nil {
				log.Printf("Background Task Error: Failed to cleanup expired sessions: %v", err)
			}
		}
	}
}

func isSameSessionContext(session repository.Session, userAgent string) bool {
	storedUserAgent := strings.TrimSpace(session.UserAgent.String)
	requestUserAgent := strings.TrimSpace(userAgent)

	// Backward compatibility: older sessions may not have a stored User-Agent.
	// Keep them refreshable and bind the new rotated session to request UA if present.
	if !session.UserAgent.Valid || storedUserAgent == "" {
		return true
	}

	if requestUserAgent == "" {
		return false
	}

	return storedUserAgent == requestUserAgent
}

func rotatedUserAgent(session repository.Session, requestUserAgent string) pgtype.Text {
	if session.UserAgent.Valid && strings.TrimSpace(session.UserAgent.String) != "" {
		return session.UserAgent
	}

	trimmed := strings.TrimSpace(requestUserAgent)
	if trimmed == "" {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: trimmed, Valid: true}
}

func (s *AuthService) refreshTokenWithRepo(
	ctx context.Context,
	repo repository.Querier,
	sessionUUID pgtype.UUID,
	userAgent string,
) (*model.AuthResponse, error) {
	oldSession, err := repo.RefreshToken(ctx, sessionUUID)
	if err != nil {
		return nil, errors.New("session not found or expired")
	}
	if !isSameSessionContext(oldSession, userAgent) {
		return nil, errors.New("session context mismatch")
	}

	// 1. Delete/Deactivate the old session to prevent reuse
	if err := repo.DeleteSession(ctx, sessionUUID); err != nil {
		return nil, errors.New("failed to invalidate old session")
	}

	// 2. Create a brand new session (Rotation)
	expiresAt := pgtype.Timestamptz{
		Time:  time.Now().Add(SessionDuration),
		Valid: true,
	}
	newSession, err := repo.CreateSession(ctx, repository.CreateSessionParams{
		UserID:     oldSession.UserID,
		DeviceInfo: oldSession.DeviceInfo,
		IpAddress:  oldSession.IpAddress,
		UserAgent:  rotatedUserAgent(oldSession, userAgent),
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		return nil, errors.New("failed to create new session during rotation")
	}

	user, err := repo.GetUserByID(ctx, newSession.UserID)
	if err != nil {
		return nil, errors.New("user not found")
	}

	newSessionID := utils.UUIDToString(newSession.ID)
	token, err := utils.GenerateTokenWithSession(
		utils.UUIDToString(user.ID),
		user.Username,
		newSessionID,
		s.jwtSecret,
	)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token:     token,
		SessionID: newSessionID,
		User: model.User{
			ID:        utils.UUIDToString(user.ID),
			Username:  user.Username,
			Email:     user.Email,
			Avatar:    user.Avatar.String,
			CreatedAt: user.CreatedAt.Time,
		},
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, sessionID string, userAgent string) (*model.AuthResponse, error) {
	sessionUUID, err := utils.StringToUUID(sessionID)
	if err != nil {
		return nil, errors.New("invalid session ID")
	}

	// Session rotation must be atomic to avoid deleting old session
	// without successfully issuing a replacement.
	if s.txBeginner != nil {
		var response *model.AuthResponse
		if err := runInTx(ctx, s.txBeginner, func(txQueries *repository.Queries) error {
			var refreshErr error
			response, refreshErr = s.refreshTokenWithRepo(ctx, txQueries, sessionUUID, userAgent)
			return refreshErr
		}); err != nil {
			return nil, err
		}
		return response, nil
	}

	return s.refreshTokenWithRepo(ctx, s.repo, sessionUUID, userAgent)
}
