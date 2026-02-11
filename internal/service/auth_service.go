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
	// 1. Hash Password
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
		Phone:          "", // Optional
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
