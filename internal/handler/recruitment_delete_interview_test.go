package handler_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"hr-backend/internal/handler"
	"hr-backend/internal/repository"
	"hr-backend/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestDeleteInterview_GetInterviewNoRows_ReturnsNotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		CheckRecruiterOrAdminFunc: func(ctx context.Context, userID pgtype.UUID) (pgtype.Bool, error) {
			return pgtype.Bool{Bool: true, Valid: true}, nil
		},
		GetInterviewFunc: func(ctx context.Context, id pgtype.UUID) (repository.Interview, error) {
			return repository.Interview{}, pgx.ErrNoRows
		},
	}

	r := newDeleteInterviewRouter(mockRepo)
	req, _ := http.NewRequest("DELETE", "/recruitment/interviews/11111111-1111-1111-1111-111111111111", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestDeleteInterview_GetInterviewDBError_ReturnsInternalServerError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		CheckRecruiterOrAdminFunc: func(ctx context.Context, userID pgtype.UUID) (pgtype.Bool, error) {
			return pgtype.Bool{Bool: true, Valid: true}, nil
		},
		GetInterviewFunc: func(ctx context.Context, id pgtype.UUID) (repository.Interview, error) {
			return repository.Interview{}, errors.New("db timeout")
		},
	}

	r := newDeleteInterviewRouter(mockRepo)
	req, _ := http.NewRequest("DELETE", "/recruitment/interviews/11111111-1111-1111-1111-111111111111", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestDeleteInterview_NonPendingStatus_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		CheckRecruiterOrAdminFunc: func(ctx context.Context, userID pgtype.UUID) (pgtype.Bool, error) {
			return pgtype.Bool{Bool: true, Valid: true}, nil
		},
		GetInterviewFunc: func(ctx context.Context, id pgtype.UUID) (repository.Interview, error) {
			return repository.Interview{Status: "COMPLETED"}, nil
		},
	}

	r := newDeleteInterviewRouter(mockRepo)
	req, _ := http.NewRequest("DELETE", "/recruitment/interviews/11111111-1111-1111-1111-111111111111", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestDeleteInterview_Success_ReturnsOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	deleteCalled := false
	mockRepo := &mocks.MockQuerier{
		CheckRecruiterOrAdminFunc: func(ctx context.Context, userID pgtype.UUID) (pgtype.Bool, error) {
			return pgtype.Bool{Bool: true, Valid: true}, nil
		},
		GetInterviewFunc: func(ctx context.Context, id pgtype.UUID) (repository.Interview, error) {
			return repository.Interview{Status: "PENDING"}, nil
		},
		DeleteInterviewFunc: func(ctx context.Context, id pgtype.UUID) (int64, error) {
			deleteCalled = true
			return 1, nil
		},
	}

	r := newDeleteInterviewRouter(mockRepo)
	req, _ := http.NewRequest("DELETE", "/recruitment/interviews/11111111-1111-1111-1111-111111111111", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", w.Code, w.Body.String())
	}
	if !deleteCalled {
		t.Fatal("expected DeleteInterview to be called")
	}
}

func TestDeleteInterview_DeleteAffectsNoRows_ReturnsBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		CheckRecruiterOrAdminFunc: func(ctx context.Context, userID pgtype.UUID) (pgtype.Bool, error) {
			return pgtype.Bool{Bool: true, Valid: true}, nil
		},
		GetInterviewFunc: func(ctx context.Context, id pgtype.UUID) (repository.Interview, error) {
			return repository.Interview{Status: "PENDING"}, nil
		},
		DeleteInterviewFunc: func(ctx context.Context, id pgtype.UUID) (int64, error) {
			return 0, nil
		},
	}

	r := newDeleteInterviewRouter(mockRepo)
	req, _ := http.NewRequest("DELETE", "/recruitment/interviews/11111111-1111-1111-1111-111111111111", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d body=%s", w.Code, w.Body.String())
	}
}

func newDeleteInterviewRouter(mockRepo *mocks.MockQuerier) *gin.Engine {
	h := handler.NewRecruitmentHandler(mockRepo)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", "00000000-0000-0000-0000-000000000010")
		c.Next()
	})
	r.DELETE("/recruitment/interviews/:id", h.DeleteInterview)
	return r
}
