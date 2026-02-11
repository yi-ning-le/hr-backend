package service

import (
	"context"
	"errors"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

type AuthService struct {
	repo      repository.Querier
	jwtSecret string
}

func NewAuthService(repo repository.Querier, jwtSecret string) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Register(ctx context.Context, input model.RegisterInput) (*model.User, error) {
	// 1. Hash Password
	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	// 2. Create User
	// Note: We should check if username/email exists, but the DB unique constraint handles it too.
	// For better UX, we could do a read before write. Relying on DB constraint for now.

	params := repository.CreateUserParams{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: hashedPassword,
		Avatar:       pgtype.Text{Valid: false}, // Default null
	}

	user, err := s.repo.CreateUser(ctx, params)
	if err != nil {
		// Basic duplicate check logic (in a real app, parse pg error)
		return nil, err
	}

	// 3. Auto-create Employee record to link with User
	// Use username as placeholder for First/Last name if not provided (RegisterInput might need expansion later)
	// For now, split username or just use it.
	firstName := input.Username
	lastName := "User"

	// Create timestamp for join date (now)
	now := pgtype.Timestamptz{
		Time:  utils.Now(),
		Valid: true,
	}

	empParams := repository.CreateEmployeeParams{
		FirstName:      firstName,
		LastName:       lastName,
		Email:          input.Email,
		Phone:          "", // Optional
		Department:     "Unassigned",
		Position:       "New Hire",
		Status:         "Active",
		EmploymentType: "FullTime",
		JoinDate:       now,
		UserID:         user.ID, // Link to the new user!
		ManagerID:      pgtype.UUID{Valid: false},
	}

	_, err = s.repo.CreateEmployee(ctx, empParams)
	if err != nil {
		_ = s.repo.DeleteUser(ctx, user.ID)
		return nil, errors.New("failed to create linked employee record: " + err.Error())
	}

	return &model.User{
		ID:        utils.UUIDToString(user.ID),
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Time,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, input model.LoginInput) (*model.AuthResponse, error) {
	// 1. Find User
	user, err := s.repo.GetUserByUsername(ctx, input.Username)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// 2. Check Password
	if !utils.CheckPasswordHash(input.Password, user.PasswordHash) {
		return nil, errors.New("invalid credentials")
	}

	// 3. Generate Token
	token, err := utils.GenerateToken(utils.UUIDToString(user.ID), user.Username, s.jwtSecret)
	if err != nil {
		return nil, err
	}

	return &model.AuthResponse{
		Token: token,
		User: model.User{
			ID:        utils.UUIDToString(user.ID),
			Username:  user.Username,
			Email:     user.Email,
			Avatar:    user.Avatar.String,
			CreatedAt: user.CreatedAt.Time,
		},
	}, nil
}
