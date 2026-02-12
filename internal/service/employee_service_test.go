package service_test

import (
	"context"
	"testing"
	"time"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/service"
	"hr-backend/test/mocks"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestCreateEmployee_AutoCreatesUser(t *testing.T) {
	var createdUserParams repository.CreateUserParams
	createUserCalled := false

	mockRepo := &mocks.MockQuerier{
		CreateUserFunc: func(ctx context.Context, arg repository.CreateUserParams) (repository.User, error) {
			createUserCalled = true
			createdUserParams = arg
			return repository.User{
				ID:       pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
				Username: arg.Username,
				Email:    arg.Email,
			}, nil
		},
		CreateEmployeeFunc: func(ctx context.Context, arg repository.CreateEmployeeParams) (repository.Employee, error) {
			return repository.Employee{
				ID:             pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				FirstName:      arg.FirstName,
				LastName:       arg.LastName,
				Email:          arg.Email,
				Phone:          arg.Phone,
				Department:     arg.Department,
				Position:       arg.Position,
				Status:         arg.Status,
				EmploymentType: arg.EmploymentType,
				EmployeeType:   arg.EmployeeType,
				JoinDate:       arg.JoinDate,
				UserID:         arg.UserID,
			}, nil
		},
	}

	svc := service.NewEmployeeService(mockRepo)

	input := model.EmployeeInput{
		FirstName:  "张",
		LastName:   "三",
		Email:      "zhangsan@example.com",
		Phone:      "13800138000",
		Department: "技术部",
		Position:   "高级工程师",
		JoinDate:   time.Now(),
	}

	employee, err := svc.CreateEmployee(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify user was created automatically
	if !createUserCalled {
		t.Fatal("expected CreateUser to be called, but it was not")
	}

	// Verify user params: username = email, default password hash
	if createdUserParams.Username != input.Email {
		t.Errorf("expected username %s, got %s", input.Email, createdUserParams.Username)
	}
	if createdUserParams.Email != input.Email {
		t.Errorf("expected email %s, got %s", input.Email, createdUserParams.Email)
	}
	// Password should be hashed (not empty)
	if createdUserParams.PasswordHash == "" {
		t.Error("expected password hash to be set")
	}

	// Verify employee was created with correct user link
	if employee.FirstName != input.FirstName {
		t.Errorf("expected firstName %s, got %s", input.FirstName, employee.FirstName)
	}
	if employee.TemporaryPassword == "" {
		t.Error("expected temporaryPassword to be returned for first login delivery")
	}
	if employee.Status != "Active" {
		t.Errorf("expected default status Active, got %s", employee.Status)
	}
	if employee.EmploymentType != "FullTime" {
		t.Errorf("expected default employmentType FullTime, got %s", employee.EmploymentType)
	}
	// Default employee type should be EMPLOYEE
	if employee.EmployeeType != "EMPLOYEE" {
		t.Errorf("expected default employeeType EMPLOYEE, got %s", employee.EmployeeType)
	}
}

func TestCreateEmployee_RollsBackUserWhenEmployeeCreateFails(t *testing.T) {
	createdUserID := pgtype.UUID{Bytes: [16]byte{7}, Valid: true}
	deleteCalled := false

	mockRepo := &mocks.MockQuerier{
		CreateUserFunc: func(ctx context.Context, arg repository.CreateUserParams) (repository.User, error) {
			return repository.User{
				ID: createdUserID,
			}, nil
		},
		CreateEmployeeFunc: func(ctx context.Context, arg repository.CreateEmployeeParams) (repository.Employee, error) {
			return repository.Employee{}, assert.AnError
		},
		DeleteUserFunc: func(ctx context.Context, id pgtype.UUID) error {
			deleteCalled = true
			if id != createdUserID {
				t.Errorf("expected deleted user id %v, got %v", createdUserID, id)
			}
			return nil
		},
	}

	svc := service.NewEmployeeService(mockRepo)

	input := model.EmployeeInput{
		FirstName:  "回滚",
		LastName:   "测试",
		Email:      "rollback@example.com",
		Phone:      "13800138099",
		Department: "技术部",
		Position:   "工程师",
		JoinDate:   time.Now(),
	}

	employee, err := svc.CreateEmployee(context.Background(), input)
	if err == nil {
		t.Fatal("expected error when CreateEmployee fails")
	}
	if employee != nil {
		t.Fatal("expected nil employee on failure")
	}
	if !deleteCalled {
		t.Fatal("expected DeleteUser rollback to be called")
	}
}

func TestCreateEmployee_HR(t *testing.T) {
	mockRepo := &mocks.MockQuerier{
		CreateUserFunc: func(ctx context.Context, arg repository.CreateUserParams) (repository.User, error) {
			return repository.User{
				ID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
			}, nil
		},
		CreateEmployeeFunc: func(ctx context.Context, arg repository.CreateEmployeeParams) (repository.Employee, error) {
			return repository.Employee{
				ID:           pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				EmployeeType: arg.EmployeeType,
			}, nil
		},
	}

	svc := service.NewEmployeeService(mockRepo)

	input := model.EmployeeInput{
		FirstName:    "李",
		LastName:     "四",
		Email:        "lisi@example.com",
		Phone:        "13800138001",
		Department:   "人力资源部",
		Position:     "HR",
		JoinDate:     time.Now(),
		EmployeeType: "HR",
	}

	employee, err := svc.CreateEmployee(context.Background(), input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if employee.EmployeeType != "HR" {
		t.Errorf("expected employeeType HR, got %s", employee.EmployeeType)
	}
}

func TestGetEmployee(t *testing.T) {
	employeeIDStr := "01010101-0101-0101-0101-010101010101"
	var employeeIDUUID pgtype.UUID
	if err := employeeIDUUID.Scan(employeeIDStr); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		GetEmployeeFunc: func(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{
				ID:             employeeIDUUID,
				FirstName:      "Test",
				LastName:       "User",
				Email:          "test@example.com",
				Phone:          "1234567890",
				Department:     "Engineering",
				Position:       "Developer",
				Status:         "Active",
				EmploymentType: "FullTime",
				JoinDate:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	svc := service.NewEmployeeService(mockRepo)

	employee, err := svc.GetEmployee(context.Background(), employeeIDStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if employee.FirstName != "Test" {
		t.Errorf("expected firstName Test, got %s", employee.FirstName)
	}
	if employee.ID != employeeIDStr {
		t.Errorf("expected ID %s, got %s", employeeIDStr, employee.ID)
	}
}

func TestListEmployees(t *testing.T) {
	mockRepo := &mocks.MockQuerier{
		ListEmployeesFunc: func(ctx context.Context, arg repository.ListEmployeesParams) ([]repository.Employee, error) {
			return []repository.Employee{
				{
					ID:             pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
					FirstName:      "Alice",
					LastName:       "Smith",
					Email:          "alice@example.com",
					Status:         "Active",
					EmploymentType: "FullTime",
					JoinDate:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				},
				{
					ID:             pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
					FirstName:      "Bob",
					LastName:       "Jones",
					Email:          "bob@example.com",
					Status:         "Active",
					EmploymentType: "PartTime",
					JoinDate:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				},
			}, nil
		},
		CountEmployeesFunc: func(ctx context.Context, arg repository.CountEmployeesParams) (int64, error) {
			return 2, nil
		},
	}

	svc := service.NewEmployeeService(mockRepo)

	result, err := svc.ListEmployees(context.Background(), "", "", "", 1, 20)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Employees) != 2 {
		t.Errorf("expected 2 employees, got %d", len(result.Employees))
	}
	if result.Total != 2 {
		t.Errorf("expected total 2, got %d", result.Total)
	}
}

func TestDeleteEmployee(t *testing.T) {
	employeeIDStr := "01010101-0101-0101-0101-010101010101"
	deleteCalled := false

	mockRepo := &mocks.MockQuerier{
		DeleteEmployeeFunc: func(ctx context.Context, id pgtype.UUID) error {
			deleteCalled = true
			return nil
		},
	}

	svc := service.NewEmployeeService(mockRepo)

	err := svc.DeleteEmployee(context.Background(), employeeIDStr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !deleteCalled {
		t.Error("expected DeleteEmployee to be called")
	}
}

func TestUpdateEmployee_PreservesUserIDWhenInputEmpty(t *testing.T) {
	employeeIDStr := "11111111-1111-1111-1111-111111111111"
	existingUserIDStr := "22222222-2222-2222-2222-222222222222"

	var employeeID, existingUserID pgtype.UUID
	if err := employeeID.Scan(employeeIDStr); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}
	if err := existingUserID.Scan(existingUserIDStr); err != nil {
		t.Fatalf("failed to scan existing user id: %v", err)
	}

	var captured repository.UpdateEmployeeParams
	mockRepo := &mocks.MockQuerier{
		GetEmployeeFunc: func(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{
				ID:     id,
				UserID: existingUserID,
			}, nil
		},
		UpdateEmployeeFunc: func(ctx context.Context, arg repository.UpdateEmployeeParams) (repository.Employee, error) {
			captured = arg
			return repository.Employee{
				ID:     arg.ID,
				UserID: arg.UserID,
			}, nil
		},
	}

	svc := service.NewEmployeeService(mockRepo)
	_, err := svc.UpdateEmployee(context.Background(), employeeIDStr, model.EmployeeInput{
		FirstName:      "A",
		LastName:       "B",
		Email:          "ab@example.com",
		Phone:          "1",
		Department:     "D",
		Position:       "P",
		Status:         "Active",
		EmploymentType: "FullTime",
		JoinDate:       time.Now(),
		UserID:         "",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assert.Equal(t, existingUserID, captured.UserID)
}
