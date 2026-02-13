package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hr-backend/internal/handler"
	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestCreateInterview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	candidateIDStr := "11111111-1111-1111-1111-111111111111"
	interviewerIDStr := "22222222-2222-2222-2222-222222222222"
	jobIDStr := "33333333-3333-3333-3333-333333333333"

	var candidateID, interviewerID, jobID pgtype.UUID
	if err := candidateID.Scan(candidateIDStr); err != nil {
		t.Fatalf("failed to scan candidate id: %v", err)
	}
	if err := interviewerID.Scan(interviewerIDStr); err != nil {
		t.Fatalf("failed to scan interviewer id: %v", err)
	}
	if err := jobID.Scan(jobIDStr); err != nil {
		t.Fatalf("failed to scan job id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		CreateInterviewFunc: func(ctx context.Context, arg repository.CreateInterviewParams) (repository.Interview, error) {
			return repository.Interview{
				ID:            pgtype.UUID{Valid: true}, // returned ID doesn't matter much for this test
				CandidateID:   arg.CandidateID,
				InterviewerID: arg.InterviewerID,
				JobID:         arg.JobID,
				ScheduledTime: arg.ScheduledTime,
				Status:        arg.Status,
				CreatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
				UpdatedAt:     pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)
	r := gin.New()
	r.POST("/recruitment/interviews", h.CreateInterview)

	input := model.CreateInterviewInput{
		CandidateID:   candidateIDStr,
		InterviewerID: interviewerIDStr,
		JobID:         jobIDStr,
		ScheduledTime: time.Now().Add(24 * time.Hour),
	}

	body, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/recruitment/interviews", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", w.Code)
	}

	var result model.Interview
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result.CandidateID != candidateIDStr {
		t.Errorf("expected candidateId %s, got %s", candidateIDStr, result.CandidateID)
	}
}

func TestGetMyInterviews(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userIDStr := "11111111-1111-1111-1111-111111111111"
	var userID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		t.Fatalf("failed to scan user id: %v", err)
	}

	employeeIDStr := "22222222-2222-2222-2222-222222222222"
	var employeeID pgtype.UUID
	if err := employeeID.Scan(employeeIDStr); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, uid pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{ID: employeeID, UserID: userID}, nil
		},
		ListInterviewsByInterviewerFunc: func(ctx context.Context, eid pgtype.UUID) ([]repository.Interview, error) {
			if eid != employeeID {
				return nil, nil
			}
			return []repository.Interview{
				{
					ID:            pgtype.UUID{Valid: true},
					InterviewerID: employeeID,
					Status:        "PENDING",
				},
			}, nil
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.GET("/recruitment/interviews/me", h.GetMyInterviews)

	req, _ := http.NewRequest("GET", "/recruitment/interviews/me", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result []model.Interview
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if len(result) != 1 {
		t.Errorf("expected 1 interview, got %d", len(result))
	}
}

func TestGetInterview(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userIDStr := "11111111-1111-1111-1111-111111111111"
	employeeIDStr := "22222222-2222-2222-2222-222222222222"
	interviewIDStr := "33333333-3333-3333-3333-333333333333"

	var userID, employeeID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		t.Fatalf("failed to scan user id: %v", err)
	}
	if err := employeeID.Scan(employeeIDStr); err != nil {
		t.Fatalf("failed to scan employee id: %v", err)
	}

	var interviewID pgtype.UUID
	if err := interviewID.Scan(interviewIDStr); err != nil {
		t.Fatalf("failed to scan interview id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
			if id != userID {
				return repository.Employee{}, errors.New("not found")
			}
			return repository.Employee{ID: employeeID, UserID: userID}, nil
		},
		GetInterviewFunc: func(ctx context.Context, id pgtype.UUID) (repository.Interview, error) {
			if id != interviewID {
				return repository.Interview{}, errors.New("not found")
			}
			return repository.Interview{
				ID:            interviewID,
				InterviewerID: employeeID,
				Status:        "PENDING",
			}, nil
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.GET("/recruitment/interviews/:id", h.GetInterview)

	req, _ := http.NewRequest("GET", "/recruitment/interviews/"+interviewIDStr, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var result model.Interview
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if result.ID != interviewIDStr {
		t.Errorf("expected id %s, got %s", interviewIDStr, result.ID)
	}
}

func TestGetInterview_ForbiddenForNonOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userIDStr := "11111111-1111-1111-1111-111111111111"
	ownerEmployeeIDStr := "22222222-2222-2222-2222-222222222222"
	requesterEmployeeIDStr := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	interviewIDStr := "33333333-3333-3333-3333-333333333333"

	var userID, ownerEmployeeID, requesterEmployeeID, interviewID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		t.Fatalf("failed to scan user id: %v", err)
	}
	if err := ownerEmployeeID.Scan(ownerEmployeeIDStr); err != nil {
		t.Fatalf("failed to scan owner employee id: %v", err)
	}
	if err := requesterEmployeeID.Scan(requesterEmployeeIDStr); err != nil {
		t.Fatalf("failed to scan requester employee id: %v", err)
	}
	if err := interviewID.Scan(interviewIDStr); err != nil {
		t.Fatalf("failed to scan interview id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{ID: requesterEmployeeID, UserID: userID}, nil
		},
		GetInterviewFunc: func(ctx context.Context, id pgtype.UUID) (repository.Interview, error) {
			return repository.Interview{
				ID:            interviewID,
				InterviewerID: ownerEmployeeID,
				Status:        "PENDING",
			}, nil
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.GET("/recruitment/interviews/:id", h.GetInterview)

	req, _ := http.NewRequest("GET", "/recruitment/interviews/"+interviewIDStr, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestGetInterview_AllowsRecruiterNonOwner(t *testing.T) {
	gin.SetMode(gin.TestMode)

	userIDStr := "11111111-1111-1111-1111-111111111111"
	ownerEmployeeIDStr := "22222222-2222-2222-2222-222222222222"
	requesterEmployeeIDStr := "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
	interviewIDStr := "33333333-3333-3333-3333-333333333333"

	var userID, ownerEmployeeID, requesterEmployeeID, interviewID pgtype.UUID
	if err := userID.Scan(userIDStr); err != nil {
		t.Fatalf("failed to scan user id: %v", err)
	}
	if err := ownerEmployeeID.Scan(ownerEmployeeIDStr); err != nil {
		t.Fatalf("failed to scan owner employee id: %v", err)
	}
	if err := requesterEmployeeID.Scan(requesterEmployeeIDStr); err != nil {
		t.Fatalf("failed to scan requester employee id: %v", err)
	}
	if err := interviewID.Scan(interviewIDStr); err != nil {
		t.Fatalf("failed to scan interview id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{ID: requesterEmployeeID, UserID: userID}, nil
		},
		GetInterviewFunc: func(ctx context.Context, id pgtype.UUID) (repository.Interview, error) {
			return repository.Interview{
				ID:            interviewID,
				InterviewerID: ownerEmployeeID,
				Status:        "PENDING",
			}, nil
		},
		CheckRecruiterRoleFunc: func(ctx context.Context, employeeID pgtype.UUID) (pgtype.UUID, error) {
			return employeeID, nil
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.GET("/recruitment/interviews/:id", h.GetInterview)

	req, _ := http.NewRequest("GET", "/recruitment/interviews/"+interviewIDStr, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
