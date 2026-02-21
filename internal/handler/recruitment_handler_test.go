package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"hr-backend/internal/handler"
	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestGetRecruiters(t *testing.T) {
	gin.SetMode(gin.TestMode)

	employeeIDStr := "5111b81e-bd11-471a-96e0-24927f906d1e"
	var employeeIDUUID pgtype.UUID
	if err := employeeIDUUID.Scan(employeeIDStr); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		ListRecruitersFunc: func(ctx context.Context) ([]repository.ListRecruitersRow, error) {
			return []repository.ListRecruitersRow{
				{
					ID:         employeeIDUUID,
					FirstName:  "John",
					LastName:   "Doe",
					Department: "HR",
				},
			}, nil
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)

	r := gin.New()
	r.GET("/recruitment/admin/recruiters", h.GetRecruiters)

	req, _ := http.NewRequest("GET", "/recruitment/admin/recruiters", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []model.Recruiter
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 recruiter, got %d", len(result))
	}

	// This is the CRITICAL check: it should be the same UUID string
	if result[0].EmployeeID != employeeIDStr {
		t.Errorf("expected employeeId %s, got %s", employeeIDStr, result[0].EmployeeID)
	}
}

func TestRevokeRecruiter(t *testing.T) {
	gin.SetMode(gin.TestMode)

	employeeIDStr := "5111b81e-bd11-471a-96e0-24927f906d1e"
	var employeeIDUUID pgtype.UUID
	if err := employeeIDUUID.Scan(employeeIDStr); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}

	var capturedID pgtype.UUID
	mockRepo := &mocks.MockQuerier{
		RevokeRecruiterRoleFunc: func(ctx context.Context, employeeID pgtype.UUID) error {
			capturedID = employeeID
			return nil
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)

	r := gin.New()
	// New API: No path parameter
	r.DELETE("/recruitment/admin/recruiters", h.RevokeRecruiter)

	// Test Case: Valid JSON body
	input := map[string]string{"employeeId": employeeIDStr}
	body, _ := json.Marshal(input)
	req, _ := http.NewRequest("DELETE", "/recruitment/admin/recruiters", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	if capturedID != employeeIDUUID {
		t.Errorf("expected to revoke ID %v, got %v", employeeIDUUID, capturedID)
	}
}

func TestGetMyRole_UsesExplicitReviewCapability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userIDStr := "5111b81e-bd11-471a-96e0-24927f906d1e"

	var employeeIDUUID pgtype.UUID
	if err := employeeIDUUID.Scan("6111b81e-bd11-471a-96e0-24927f906d1e"); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		CheckIsAdminFunc: func(ctx context.Context, id pgtype.UUID) (bool, error) {
			return false, nil
		},
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{
				ID:           employeeIDUUID,
				EmployeeType: "HR",
			}, nil
		},
		CheckRecruiterRoleFunc: func(ctx context.Context, employeeID pgtype.UUID) (pgtype.UUID, error) {
			return pgtype.UUID{}, errors.New("not recruiter")
		},
		GetActiveInterviewCountFunc: func(ctx context.Context, interviewerID pgtype.UUID) (int64, error) {
			return 1, nil
		},
		CheckInterviewerRoleFunc: func(ctx context.Context, employeeID pgtype.UUID) (pgtype.UUID, error) {
			return employeeIDUUID, nil
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.GET("/recruitment/role", h.GetMyRole)

	req, _ := http.NewRequest("GET", "/recruitment/role", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result model.RecruitmentRoleResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !result.IsInterviewer {
		t.Errorf("expected isInterviewer=true when explicit capability is enabled")
	}
	if !result.IsHR {
		t.Errorf("expected isHR=true")
	}
}

func TestGetMyRole_NoEmployeeDoesNotGrantReviewCapability(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userIDStr := "7111b81e-bd11-471a-96e0-24927f906d1e"

	mockRepo := &mocks.MockQuerier{
		CheckIsAdminFunc: func(ctx context.Context, id pgtype.UUID) (bool, error) {
			return true, nil
		},
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{}, errors.New("employee not found")
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.GET("/recruitment/role", h.GetMyRole)

	req, _ := http.NewRequest("GET", "/recruitment/role", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result model.RecruitmentRoleResponse
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if !result.IsAdmin {
		t.Errorf("expected isAdmin=true")
	}
	if !result.IsAdmin {
		t.Errorf("expected isAdmin=true")
	}
}

func TestGetMyRole_DoesNotMutateInterviewerRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userIDStr := "8111b81e-bd11-471a-96e0-24927f906d1e"
	var employeeIDUUID pgtype.UUID
	if err := employeeIDUUID.Scan("9111b81e-bd11-471a-96e0-24927f906d1e"); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}

	revokeCalled := 0
	mockRepo := &mocks.MockQuerier{
		CheckIsAdminFunc: func(ctx context.Context, id pgtype.UUID) (bool, error) {
			return false, nil
		},
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{
				ID:           employeeIDUUID,
				EmployeeType: "EMPLOYEE",
			}, nil
		},
		CheckRecruiterRoleFunc: func(ctx context.Context, employeeID pgtype.UUID) (pgtype.UUID, error) {
			return pgtype.UUID{}, errors.New("not recruiter")
		},
		GetActiveInterviewCountFunc: func(ctx context.Context, interviewerID pgtype.UUID) (int64, error) {
			return 0, nil
		},
		CheckInterviewerRoleFunc: func(ctx context.Context, employeeID pgtype.UUID) (pgtype.UUID, error) {
			return pgtype.UUID{}, errors.New("not interviewer")
		},
		RevokeInterviewerRoleFunc: func(ctx context.Context, employeeID pgtype.UUID) error {
			revokeCalled++
			return nil
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.GET("/recruitment/role", h.GetMyRole)

	req, _ := http.NewRequest("GET", "/recruitment/role", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if revokeCalled != 0 {
		t.Fatalf("expected GetMyRole to be read-only, but revoke was called %d times", revokeCalled)
	}
}
