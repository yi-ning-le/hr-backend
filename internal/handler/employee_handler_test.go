package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hr-backend/internal/handler"
	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/service"
	"hr-backend/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestListEmployeesHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

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
			}, nil
		},
		CountEmployeesFunc: func(ctx context.Context, arg repository.CountEmployeesParams) (int64, error) {
			return 1, nil
		},
	}

	svc := service.NewEmployeeService(mockRepo)
	h := handler.NewEmployeeHandler(svc)

	r := gin.New()
	r.GET("/employees", h.ListEmployees)

	req, _ := http.NewRequest("GET", "/employees", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result model.EmployeeListResult
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(result.Employees) != 1 {
		t.Errorf("expected 1 employee, got %d", len(result.Employees))
	}
}

func TestCreateEmployeeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		CreateUserFunc: func(ctx context.Context, arg repository.CreateUserParams) (repository.User, error) {
			return repository.User{
				ID: pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
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
	h := handler.NewEmployeeHandler(svc)

	r := gin.New()
	r.POST("/employees", h.CreateEmployee)

	input := model.EmployeeInput{
		FirstName:  "Test",
		LastName:   "User",
		Email:      "test@example.com",
		Phone:      "1234567890",
		Department: "Engineering",
		Position:   "Developer",
		JoinDate:   time.Now(),
	}
	body, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/employees", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestGetEmployeeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

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
				Status:         "Active",
				EmploymentType: "FullTime",
				JoinDate:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	svc := service.NewEmployeeService(mockRepo)
	h := handler.NewEmployeeHandler(svc)

	r := gin.New()
	r.GET("/employees/:id", h.GetEmployee)

	req, _ := http.NewRequest("GET", "/employees/"+employeeIDStr, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestDeleteEmployeeHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	employeeIDStr := "01010101-0101-0101-0101-010101010101"

	mockRepo := &mocks.MockQuerier{
		DeleteEmployeeFunc: func(ctx context.Context, id pgtype.UUID) error {
			return nil
		},
	}

	svc := service.NewEmployeeService(mockRepo)
	h := handler.NewEmployeeHandler(svc)

	r := gin.New()
	r.DELETE("/employees/:id", h.DeleteEmployee)

	req, _ := http.NewRequest("DELETE", "/employees/"+employeeIDStr, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}
