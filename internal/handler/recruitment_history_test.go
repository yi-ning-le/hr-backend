package handler_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hr-backend/internal/handler"
	"hr-backend/internal/repository"
	"hr-backend/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestGetCandidateHistory_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reviewerIDStr := "11111111-1111-1111-1111-111111111111"
	var userID pgtype.UUID
	err := userID.Scan(reviewerIDStr)
	assert.NoError(t, err)

	reviewerEmployeeIDStr := "33333333-3333-3333-3333-333333333333"
	var reviewerEmployeeID pgtype.UUID
	err = reviewerEmployeeID.Scan(reviewerEmployeeIDStr)
	assert.NoError(t, err)

	candidateIDStr := "22222222-2222-2222-2222-222222222222"
	var candidateID pgtype.UUID
	err = candidateID.Scan(candidateIDStr)
	assert.NoError(t, err)

	timeNow := time.Now()

	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, userIDArg pgtype.UUID) (repository.Employee, error) {
			assert.Equal(t, userID, userIDArg)
			return repository.Employee{ID: reviewerEmployeeID}, nil
		},
		// Mock GetCandidateHistoryForReviewer
		GetCandidateHistoryForReviewerFunc: func(ctx context.Context, arg repository.GetCandidateHistoryForReviewerParams) ([]repository.GetCandidateHistoryForReviewerRow, error) {
			assert.Equal(t, candidateID, arg.CandidateID)
			assert.Equal(t, reviewerEmployeeID, arg.ReviewerID)

			return []repository.GetCandidateHistoryForReviewerRow{
				{
					CandidateID:   candidateID,
					CandidateName: "John Doe",
					Status:        "interviewing",
					ReviewStatus:  "suitable",
					AppliedAt:     pgtype.Timestamptz{Time: timeNow, Valid: true},
					JobTitle:      "Software Engineer",
				},
				{
					CandidateID:   candidateID,
					CandidateName: "John Doe",
					Status:        "rejected",
					ReviewStatus:  "unsuitable",
					AppliedAt:     pgtype.Timestamptz{Time: timeNow.AddDate(0, -1, 0), Valid: true}, // 1 month ago
					JobTitle:      "Backend Developer",
				},
			}, nil
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", reviewerIDStr)
		c.Next()
	})
	r.GET("/api/candidates/:id/history", h.GetCandidateHistory)

	req, _ := http.NewRequest("GET", "/api/candidates/"+candidateIDStr+"/history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Len(t, response, 2)
	assert.Equal(t, "interviewing", response[0]["status"])
	assert.Equal(t, "suitable", response[0]["reviewStatus"])
	assert.Equal(t, "Software Engineer", response[0]["jobTitle"])

	assert.Equal(t, "rejected", response[1]["status"])
	assert.Equal(t, "unsuitable", response[1]["reviewStatus"])
	assert.Equal(t, "Backend Developer", response[1]["jobTitle"])
}

func TestGetCandidateHistory_QueryError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reviewerIDStr := "11111111-1111-1111-1111-111111111111"
	candidateIDStr := "22222222-2222-2222-2222-222222222222"

	mockRepo := &mocks.MockQuerier{
		GetEmployeeByUserIDFunc: func(ctx context.Context, userIDArg pgtype.UUID) (repository.Employee, error) {
			return repository.Employee{ID: userIDArg}, nil
		},
		GetCandidateHistoryForReviewerFunc: func(ctx context.Context, arg repository.GetCandidateHistoryForReviewerParams) ([]repository.GetCandidateHistoryForReviewerRow, error) {
			return nil, errors.New("db error")
		},
	}

	h := handler.NewRecruitmentHandler(mockRepo)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("userID", reviewerIDStr)
		c.Next()
	})
	r.GET("/api/candidates/:id/history", h.GetCandidateHistory)

	req, _ := http.NewRequest("GET", "/api/candidates/"+candidateIDStr+"/history", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}
