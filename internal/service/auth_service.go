package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

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
		Time:  time.Now().Add(24 * time.Hour),
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
