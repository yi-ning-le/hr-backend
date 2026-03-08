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

const (
	testInterviewIDStr   = "1111b81e-bd11-471a-96e0-24927f906d1e"
	testInterviewerIDStr = "2222b81e-bd11-471a-96e0-24927f906d1e"
	testUserIDStr        = "3333b81e-bd11-471a-96e0-24927f906d1e"
)

func TestUpdateInterviewStatus_CompletedSuccess(t *testing.T) {
	mockRepo := baseUpdateInterviewStatusMock(t)
	mockRepo.UpdateInterviewStatusFunc = func(ctx context.Context, arg repository.UpdateInterviewStatusParams) (repository.Interview, error) {
		return repository.Interview{
			ID:               arg.ID,
			Status:           arg.Status,
			CreatedAt:        pgtype.Timestamptz{Time: time.Now(), Valid: true},
			ScheduledTime:    pgtype.Timestamptz{Time: time.Now().Add(24 * time.Hour), Valid: true},
			ScheduledEndTime: pgtype.Timestamptz{Time: time.Now().Add(25 * time.Hour), Valid: true},
		}, nil
	}
	mockRepo.CreateCandidateCommentFunc = func(ctx context.Context, arg repository.CreateCandidateCommentParams) (repository.CandidateComment, error) {
		return repository.CandidateComment{}, nil
	}
	mockRepo.DeleteNotificationsBySubjectIDAndEventTypeFunc = func(ctx context.Context, arg repository.DeleteNotificationsBySubjectIDAndEventTypeParams) error {
		return nil
	}

	router := setupUpdateInterviewStatusRouter(mockRepo)

	body := map[string]string{
		"status":  "COMPLETED",
		"result":  "PASS",
		"comment": "Good candidate",
	}
	resp := performUpdateInterviewStatusRequest(t, router, body)

	if resp.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d. body=%s", resp.Code, resp.Body.String())
	}
}

func TestUpdateInterviewStatus_CompletedRequiresResult(t *testing.T) {
	updateCalled := false
	mockRepo := baseUpdateInterviewStatusMock(t)
	mockRepo.UpdateInterviewStatusFunc = func(ctx context.Context, arg repository.UpdateInterviewStatusParams) (repository.Interview, error) {
		updateCalled = true
		return repository.Interview{}, nil
	}

	router := setupUpdateInterviewStatusRouter(mockRepo)

	body := map[string]string{
		"status": "COMPLETED",
	}
	resp := performUpdateInterviewStatusRequest(t, router, body)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d. body=%s", resp.Code, resp.Body.String())
	}
	if updateCalled {
		t.Fatalf("expected no status update when result is missing")
	}
}

func TestUpdateInterviewStatus_CancelledRejectsResultAndComment(t *testing.T) {
	updateCalled := false
	mockRepo := baseUpdateInterviewStatusMock(t)
	mockRepo.UpdateInterviewStatusFunc = func(ctx context.Context, arg repository.UpdateInterviewStatusParams) (repository.Interview, error) {
		updateCalled = true
		return repository.Interview{}, nil
	}

	router := setupUpdateInterviewStatusRouter(mockRepo)

	body := map[string]string{
		"status":  "CANCELLED",
		"result":  "FAIL",
		"comment": "N/A",
	}
	resp := performUpdateInterviewStatusRequest(t, router, body)

	if resp.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d. body=%s", resp.Code, resp.Body.String())
	}
	if updateCalled {
		t.Fatalf("expected no status update when cancelled payload contains result/comment")
	}
}

func baseUpdateInterviewStatusMock(t *testing.T) *mocks.MockQuerier {
	t.Helper()

	interviewID, interviewerID := mustParseTestIDs(t)
	_ = interviewID

	return &mocks.MockQuerier{
		GetInterviewFunc: func(ctx context.Context, id pgtype.UUID) (repository.Interview, error) {
			return repository.Interview{
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
			return pgtype.UUID{Valid: false}, nil
		},
	}
}

func setupUpdateInterviewStatusRouter(mockRepo *mocks.MockQuerier) *gin.Engine {
	gin.SetMode(gin.TestMode)

	h := handler.NewRecruitmentHandler(mockRepo)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", testUserIDStr)
		c.Next()
	})
	r.PATCH("/recruitment/interviews/:id/status", h.UpdateInterviewStatus)
	return r
}

func performUpdateInterviewStatusRequest(
	t *testing.T,
	router *gin.Engine,
	body map[string]string,
) *httptest.ResponseRecorder {
	t.Helper()

	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}

	req, err := http.NewRequest(
		http.MethodPatch,
		"/recruitment/interviews/"+testInterviewIDStr+"/status",
		bytes.NewBuffer(payload),
	)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	return w
}

func mustParseTestIDs(t *testing.T) (pgtype.UUID, pgtype.UUID) {
	t.Helper()

	var interviewID pgtype.UUID
	if err := interviewID.Scan(testInterviewIDStr); err != nil {
		t.Fatalf("failed to parse interview id: %v", err)
	}

	var interviewerID pgtype.UUID
	if err := interviewerID.Scan(testInterviewerIDStr); err != nil {
		t.Fatalf("failed to parse interviewer id: %v", err)
	}

	return interviewID, interviewerID
}
