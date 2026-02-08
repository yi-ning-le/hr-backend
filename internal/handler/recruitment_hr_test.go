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

func TestGetHRs(t *testing.T) {
	gin.SetMode(gin.TestMode)

	employeeIDStr := "5111b81e-bd11-471a-96e0-24927f906d1e"
	var employeeIDUUID pgtype.UUID
	employeeIDUUID.Scan(employeeIDStr)

	mockRepo := &mocks.MockQuerier{
		ListHRsFunc: func(ctx context.Context) ([]repository.ListHRsRow, error) {
			return []repository.ListHRsRow{
				{
					ID:         employeeIDUUID,
					FirstName:  "Jane",
					LastName:   "HR",
					Department: "Human Resources",
				},
			}, nil
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)

	r := gin.New()
	r.GET("/recruitment/admin/hrs", h.GetHRs)

	req, _ := http.NewRequest("GET", "/recruitment/admin/hrs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var result []model.HREmployee
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 HR, got %d", len(result))
	}

	// This is the CRITICAL check: it should be the same UUID string
	if result[0].EmployeeID != employeeIDStr {
		t.Errorf("expected employeeId %s, got %s", employeeIDStr, result[0].EmployeeID)
	}
}

func TestAssignHR(t *testing.T) {
	gin.SetMode(gin.TestMode)

	employeeIDStr := "5111b81e-bd11-471a-96e0-24927f906d1e"
	var employeeIDUUID pgtype.UUID
	employeeIDUUID.Scan(employeeIDStr)

	var capturedID pgtype.UUID
	mockRepo := &mocks.MockQuerier{
		AssignHRRoleFunc: func(ctx context.Context, id pgtype.UUID) error {
			capturedID = id
			return nil
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)

	r := gin.New()
	r.POST("/recruitment/admin/hrs", h.AssignHR)

	input := map[string]string{"employeeId": employeeIDStr}
	body, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/recruitment/admin/hrs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}

	if capturedID != employeeIDUUID {
		t.Errorf("expected to assign ID %v, got %v", employeeIDUUID, capturedID)
	}
}

func TestRevokeHR(t *testing.T) {
	gin.SetMode(gin.TestMode)

	employeeIDStr := "5111b81e-bd11-471a-96e0-24927f906d1e"
	var employeeIDUUID pgtype.UUID
	employeeIDUUID.Scan(employeeIDStr)

	var capturedID pgtype.UUID
	mockRepo := &mocks.MockQuerier{
		RevokeHRRoleFunc: func(ctx context.Context, id pgtype.UUID) error {
			capturedID = id
			return nil
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)

	r := gin.New()
	r.DELETE("/recruitment/admin/hrs", h.RevokeHR)

	input := map[string]string{"employeeId": employeeIDStr}
	body, _ := json.Marshal(input)
	req, _ := http.NewRequest("DELETE", "/recruitment/admin/hrs", bytes.NewBuffer(body))
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
