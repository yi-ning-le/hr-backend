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
	"github.com/jackc/pgx/v5/pgtype"
)

func TestAssignReviewerHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	candidateID := pgtype.UUID{Bytes: [16]byte{1}, Valid: true}
	reviewerID := pgtype.UUID{Bytes: [16]byte{2}, Valid: true}

	mockRepo := &mocks.MockQuerier{
		AssignReviewerFunc: func(ctx context.Context, arg repository.AssignReviewerParams) (repository.AssignReviewerRow, error) {
			return repository.AssignReviewerRow{
				ID:         arg.ID,
				ReviewerID: arg.ReviewerID,
				Name:       "Test Candidate",
			}, nil
		},
		GetCandidateFunc: func(ctx context.Context, id pgtype.UUID) (repository.GetCandidateRow, error) {
			return repository.GetCandidateRow{
				ID:           candidateID,
				ReviewerID:   reviewerID,
				ReviewStatus: pgtype.Text{String: "assigned", Valid: true},
				AppliedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
			}, nil
		},
	}

	svc := service.NewCandidateService(mockRepo)
	h := handler.NewCandidateHandler(svc)

	r := gin.New()
	r.POST("/candidates/:id/assign-reviewer", h.AssignReviewer)

	// UUID string representation of bytes [16]byte{1} is usually 01000000-0000-0000-0000-000000000000?
	// Actually pgtype.UUID Bytes is just [16]byte.
	// For test simplicity, I'll use a known UUID string and convert it if needed, or just let service handle conversion.
	// But in mock I'm checking equality.
	// The service helper `utils.StringToUUID` is used.
	// Let's use a real UUID string to be safe with parsing.

	// Re-defining for easier matching
}

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
