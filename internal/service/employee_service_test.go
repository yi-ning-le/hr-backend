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
)

func TestCreateEmployee(t *testing.T) {
	mockRepo := &mocks.MockQuerier{
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
				JoinDate:       arg.JoinDate,
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

	if employee.FirstName != input.FirstName {
		t.Errorf("expected firstName %s, got %s", input.FirstName, employee.FirstName)
	}
	if employee.Status != "Active" {
		t.Errorf("expected default status Active, got %s", employee.Status)
	}
	if employee.EmploymentType != "FullTime" {
		t.Errorf("expected default employmentType FullTime, got %s", employee.EmploymentType)
	}
}

func TestGetEmployee(t *testing.T) {
	employeeIDStr := "01010101-0101-0101-0101-010101010101"
	var employeeIDUUID pgtype.UUID
	employeeIDUUID.Scan(employeeIDStr)

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
