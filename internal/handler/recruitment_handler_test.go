package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
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
	employeeIDUUID.Scan(employeeIDStr)

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
	employeeIDUUID.Scan(employeeIDStr)

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
