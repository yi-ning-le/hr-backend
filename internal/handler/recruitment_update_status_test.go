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
	"hr-backend/internal/repository"
	"hr-backend/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestUpdateInterviewStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Setup mock data
	interviewIDStr := "1111b81e-bd11-471a-96e0-24927f906d1e"
	var interviewID pgtype.UUID
	if err := interviewID.Scan(interviewIDStr); err != nil {
		t.Fatalf("failed to scan interviewID: %v", err)
	}

	interviewerIDStr := "2222b81e-bd11-471a-96e0-24927f906d1e"
	var interviewerID pgtype.UUID
	if err := interviewerID.Scan(interviewerIDStr); err != nil {
		t.Fatalf("failed to scan interviewerID: %v", err)
	}

	userIDStr := "3333b81e-bd11-471a-96e0-24927f906d1e"
	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		t.Fatalf("failed to scan userID: %v", err)
	}

	// Mock Employee Profile for the user
	employeeIDStr := "4444b81e-bd11-471a-96e0-24927f906d1e"
	var employeeID pgtype.UUID
	if err := employeeID.Scan(employeeIDStr); err != nil {
		t.Fatalf("failed to scan employeeID: %v", err)
	}

	// Mock repository behavior
	mockRepo := &mocks.MockQuerier{
		GetInterviewFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetInterviewRow, error) {
			// Retrieve the interview
			return repository.GetInterviewRow{
				ID:            id,
				InterviewerID: interviewerID,
				Status:        "PENDING",
			}, nil
		},
		GetEmployeeByUserIDFunc: func(ctx context.Context, uid pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{
				ID:     interviewerID,
				UserID: uid,
			}, nil
		},
		CheckRecruiterRoleFunc: func(ctx context.Context, eid pgtype.UUID) (pgtype.UUID, error) {
			// Not a recruiter
			return pgtype.UUID{Valid: false}, nil
		},
		UpdateInterviewStatusFunc: func(ctx context.Context, arg repository.UpdateInterviewStatusParams) (repository.Interview, error) {
			return repository.Interview{
				ID:               arg.ID,
				Status:           arg.Status,
				CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
				ScheduledTime:    pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
				ScheduledEndTime: pgtype.Timestamptz{Time: time.Now().Add(25 * time.Hour), Valid: true},
			}, nil
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)
	r := gin.New()

	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})

	// Register the route (Simulating main.go)
	r.PATCH("/recruitment/interviews/:id/status", h.UpdateInterviewStatus)

	// Input payload
	input := map[string]string{
		"status": "COMPLETED",
	}
	body, _ := json.Marshal(input)

	req, _ := http.NewRequest("PATCH", "/recruitment/interviews/"+interviewIDStr+"/status", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d. Body: %s", w.Code, w.Body.String())
	}
}
