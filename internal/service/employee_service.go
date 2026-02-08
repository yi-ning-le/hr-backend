package service

import (
	"context"
	"fmt"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/utils"

	"github.com/jackc/pgx/v5/pgtype"
)

const defaultPassword = "password123"

type EmployeeService struct {
	repo repository.Querier
}

func NewEmployeeService(repo repository.Querier) *EmployeeService {
	return &EmployeeService{repo: repo}
}

func (s *EmployeeService) CreateEmployee(ctx context.Context, input model.EmployeeInput) (*model.Employee, error) {
	// Set defaults
	status := input.Status
	if status == "" {
		status = "Active"
	}
	employmentType := input.EmploymentType
	if employmentType == "" {
		employmentType = "FullTime"
	}
	employeeType := input.EmployeeType
	if employeeType == "" {
		employeeType = "EMPLOYEE"
	}

	// 1. Auto-create user account for this employee
	hashedPassword, err := utils.HashPassword(defaultPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	userParams := repository.CreateUserParams{
		Username:     input.Email,
		Email:        input.Email,
		PasswordHash: hashedPassword,
		Avatar:       pgtype.Text{Valid: false},
	}

	user, err := s.repo.CreateUser(ctx, userParams)
	if err != nil {
		return nil, fmt.Errorf("failed to create user account: %w", err)
	}

	// 2. Parse manager ID if provided
	var managerID pgtype.UUID
	if input.ManagerID != "" {
		managerID, err = utils.StringToUUID(input.ManagerID)
		if err != nil {
			return nil, err
		}
	}

	// 3. Create employee linked to auto-created user
	params := repository.CreateEmployeeParams{
		FirstName:      input.FirstName,
		LastName:       input.LastName,
		Email:          input.Email,
		Phone:          input.Phone,
		Department:     input.Department,
		Position:       input.Position,
		Status:         status,
		EmploymentType: employmentType,
		EmployeeType:   employeeType,
		JoinDate:       pgtype.Timestamptz{Time: input.JoinDate, Valid: true},
		ManagerID:      managerID,
		UserID:         user.ID, // Link to auto-created user
	}

	emp, err := s.repo.CreateEmployee(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapEmployeeToModel(emp), nil
}

func (s *EmployeeService) GetEmployee(ctx context.Context, id string) (*model.Employee, error) {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	emp, err := s.repo.GetEmployee(ctx, uuid)
	if err != nil {
		return nil, err
	}

	return mapEmployeeToModel(emp), nil
}

func (s *EmployeeService) ListEmployees(ctx context.Context, status, department, search string, page, limit int) (*model.EmployeeListResult, error) {
	if limit <= 0 {
		limit = 20
	}
	if page <= 0 {
		page = 1
	}
	offset := (page - 1) * limit

	params := repository.ListEmployeesParams{
		Status:     pgtype.Text{String: status, Valid: status != ""},
		Department: pgtype.Text{String: department, Valid: department != ""},
		Search:     pgtype.Text{String: search, Valid: search != ""},
		LimitVal:   int32(limit),
		OffsetVal:  int32(offset),
	}

	employees, err := s.repo.ListEmployees(ctx, params)
	if err != nil {
		return nil, err
	}

	countParams := repository.CountEmployeesParams{
		Status:     pgtype.Text{String: status, Valid: status != ""},
		Department: pgtype.Text{String: department, Valid: department != ""},
		Search:     pgtype.Text{String: search, Valid: search != ""},
	}
	total, err := s.repo.CountEmployees(ctx, countParams)
	if err != nil {
		return nil, err
	}

	result := make([]model.Employee, len(employees))
	for i, e := range employees {
		result[i] = *mapEmployeeToModel(e)
	}

	return &model.EmployeeListResult{
		Employees: result,
		Total:     total,
		Page:      page,
		Limit:     limit,
	}, nil
}

func (s *EmployeeService) UpdateEmployee(ctx context.Context, id string, input model.EmployeeInput) (*model.Employee, error) {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	var managerID pgtype.UUID
	if input.ManagerID != "" {
		managerID, err = utils.StringToUUID(input.ManagerID)
		if err != nil {
			return nil, err
		}
	}

	var userID pgtype.UUID
	if input.UserID != "" {
		userID, err = utils.StringToUUID(input.UserID)
		if err != nil {
			return nil, err
		}
	}

	params := repository.UpdateEmployeeParams{
		ID:             uuid,
		FirstName:      input.FirstName,
		LastName:       input.LastName,
		Email:          input.Email,
		Phone:          input.Phone,
		Department:     input.Department,
		Position:       input.Position,
		Status:         input.Status,
		EmploymentType: input.EmploymentType,
		JoinDate:       pgtype.Timestamptz{Time: input.JoinDate, Valid: true},
		ManagerID:      managerID,
		UserID:         userID,
	}

	emp, err := s.repo.UpdateEmployee(ctx, params)
	if err != nil {
		return nil, err
	}

	return mapEmployeeToModel(emp), nil
}

func (s *EmployeeService) DeleteEmployee(ctx context.Context, id string) error {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return err
	}
	return s.repo.DeleteEmployee(ctx, uuid)
}

func mapEmployeeToModel(e repository.Employee) *model.Employee {
	return &model.Employee{
		ID:             utils.UUIDToString(e.ID),
		FirstName:      e.FirstName,
		LastName:       e.LastName,
		Email:          e.Email,
		Phone:          e.Phone,
		Department:     e.Department,
		Position:       e.Position,
		Status:         e.Status,
		EmploymentType: e.EmploymentType,
		EmployeeType:   e.EmployeeType,
		JoinDate:       e.JoinDate.Time,
		ManagerID:      utils.UUIDToString(e.ManagerID),
		UserID:         utils.UUIDToString(e.UserID),
	}
}
