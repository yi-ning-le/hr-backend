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
	"hr-backend/internal/service"
	"hr-backend/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestAssignReviewerHandler_realUUID(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cIDStr := "00000000-0000-0000-0000-000000000001"
	rIDStr := "00000000-0000-0000-0000-000000000002"

	mockRepo := &mocks.MockQuerier{
		AssignReviewerFunc: func(ctx context.Context, arg repository.AssignReviewerParams) (repository.AssignReviewerRow, error) {
			// verification logic
			return repository.AssignReviewerRow{
				ID:              pgtype.UUID{Bytes: [16]byte{15: 1}, Valid: true}, // approximate match
				ReviewerID:      pgtype.UUID{Bytes: [16]byte{15: 2}, Valid: true},
				AppliedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				AppliedJobTitle: "Software Engineer",
				Name:            "Test Candidate",
			}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{
				ID:         pgtype.UUID{Bytes: [16]byte{15: 1}, Valid: true},
				ReviewerID: pgtype.UUID{Bytes: [16]byte{15: 2}, Valid: true},
				AppliedAt:  pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.POST("/candidates/:id/assign-reviewer", h.AssignReviewer)

	reqBody := map[string]string{
		"reviewerId": rIDStr,
	}
	body, _ := json.Marshal(reqBody)

	req, _ := http.NewRequest("POST", "/candidates/"+cIDStr+"/assign-reviewer", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestSubmitReviewHandler_InvalidReviewStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	submitCalled := false
	mockRepo := &mocks.MockQuerier{
		SubmitReviewFunc: func(ctx context.Context, arg repository.SubmitReviewParams) (repository.SubmitReviewRow, error) {
			submitCalled = true
			return repository.SubmitReviewRow{}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.POST("/candidates/:id/review", h.SubmitReview)

	body, _ := json.Marshal(map[string]string{
		"reviewStatus": "unexpected",
		"reviewNote":   "test",
	})
	req, _ := http.NewRequest(
		"POST",
		"/candidates/00000000-0000-0000-0000-000000000001/review",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
	if submitCalled {
		t.Error("expected SubmitReview not to be called for invalid reviewStatus")
	}
}

func TestSubmitReviewHandler_NormalizesReviewStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)

	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	userIDStr := "00000000-0000-0000-0000-000000000003"
	var candidateID pgtype.UUID
	if err := candidateID.Scan(candidateIDStr); err != nil {
		t.Fatalf("failed to scan candidate id: %v", err)
	}
	var reviewerEmployeeID pgtype.UUID
	if err := reviewerEmployeeID.Scan("00000000-0000-0000-0000-000000000002"); err != nil {
		t.Fatalf("failed to scan reviewer employee id: %v", err)
	}

	submitCalled := false
	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{ID: reviewerEmployeeID}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{
				ID:         candidateID,
				ReviewerID: reviewerEmployeeID,
			}, nil
		},
		SubmitReviewFunc: func(ctx context.Context, arg repository.SubmitReviewParams) (repository.SubmitReviewRow, error) {
			submitCalled = true
			if arg.ReviewStatus.String != "suitable" {
				t.Errorf("expected normalized status suitable, got %s", arg.ReviewStatus.String)
			}
			return repository.SubmitReviewRow{
				ID:              candidateID,
				ReviewStatus:    pgtype.Text{String: arg.ReviewStatus.String, Valid: true},
				ReviewNote:      arg.ReviewNote,
				AppliedAt:       pgtype.Timestamptz{Time: time.Now(), Valid: true},
				AppliedJobTitle: "Software Engineer",
				Name:            "Test Candidate",
			}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.POST("/candidates/:id/review", h.SubmitReview)

	body, _ := json.Marshal(map[string]string{
		"reviewStatus": "  SUITABLE ",
		"reviewNote":   "strong fit",
	})
	req, _ := http.NewRequest(
		"POST",
		"/candidates/"+candidateIDStr+"/review",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if !submitCalled {
		t.Error("expected SubmitReview to be called")
	}
}

func TestSubmitReviewHandler_ReturnsForbiddenWhenNoEmployeeProfile(t *testing.T) {
	gin.SetMode(gin.TestMode)

	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	userIDStr := "00000000-0000-0000-0000-000000000003"
	submitCalled := false

	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{}, pgx.ErrNoRows
		},
		SubmitReviewFunc: func(ctx context.Context, arg repository.SubmitReviewParams) (repository.SubmitReviewRow, error) {
			submitCalled = true
			return repository.SubmitReviewRow{}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.POST("/candidates/:id/review", h.SubmitReview)

	body, _ := json.Marshal(map[string]string{
		"reviewStatus": "suitable",
		"reviewNote":   "strong fit",
	})
	req, _ := http.NewRequest(
		"POST",
		"/candidates/"+candidateIDStr+"/review",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d body=%s", w.Code, w.Body.String())
	}
	if submitCalled {
		t.Error("expected SubmitReview not to be called")
	}
}

func TestSubmitReviewHandler_ReturnsNotFoundWhenCandidateMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	candidateIDStr := "00000000-0000-0000-0000-000000000001"
	userIDStr := "00000000-0000-0000-0000-000000000003"
	submitCalled := false

	var reviewerEmployeeID pgtype.UUID
	if err := reviewerEmployeeID.Scan("00000000-0000-0000-0000-000000000002"); err != nil {
		t.Fatalf("failed to scan reviewer employee id: %v", err)
	}

	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, userID pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{ID: reviewerEmployeeID}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{}, pgx.ErrNoRows
		},
		SubmitReviewFunc: func(ctx context.Context, arg repository.SubmitReviewParams) (repository.SubmitReviewRow, error) {
			submitCalled = true
			return repository.SubmitReviewRow{}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", userIDStr)
		c.Next()
	})
	r.POST("/candidates/:id/review", h.SubmitReview)

	body, _ := json.Marshal(map[string]string{
		"reviewStatus": "suitable",
		"reviewNote":   "strong fit",
	})
	req, _ := http.NewRequest(
		"POST",
		"/candidates/"+candidateIDStr+"/review",
		bytes.NewBuffer(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d body=%s", w.Code, w.Body.String())
	}
	if submitCalled {
		t.Error("expected SubmitReview not to be called")
	}
}
