package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/service"
	"hr-backend/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestCandidateCommentHandler_ListComments(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := &mocks.MockQuerier{}
	svc := service.NewCandidateCommentService(mockRepo)
	h := NewCandidateCommentHandler(svc)

	r := gin.Default()
	r.GET("/candidates/:candidateId/comments", h.ListComments)

	candidateID := "00000000-0000-0000-0000-000000000001"

	mockRepo.ListCandidateCommentsFunc = func(ctx context.Context, id pgtype.UUID) ([]repository.ListCandidateCommentsRow, error) {
		return []repository.ListCandidateCommentsRow{
			{
				ID:           pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				CandidateID:  pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				AuthorID:     pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
				Content:      "Test comment",
				CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
				AuthorName:   "John Doe",
				AuthorAvatar: pgtype.Text{String: "avatar.png", Valid: true},
				AuthorRole:   "HR",
			},
		}, nil
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/candidates/"+candidateID+"/comments", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var comments []model.CandidateComment
	err := json.Unmarshal(w.Body.Bytes(), &comments)
	assert.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.Equal(t, "Test comment", comments[0].Content)
}

func TestCandidateCommentHandler_CreateComment(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockRepo := &mocks.MockQuerier{}
	svc := service.NewCandidateCommentService(mockRepo)
	h := NewCandidateCommentHandler(svc)

	r := gin.Default()
	// Mock auth middleware to set employeeID
	r.POST("/candidates/:candidateId/comments", func(c *gin.Context) {
		c.Set("employeeID", "00000000-0000-0000-0000-000000000002")
		c.Next()
	}, h.CreateComment)

	candidateID := "00000000-0000-0000-0000-000000000001"
	input := model.CreateCommentInput{Content: "New comment"}
	body, _ := json.Marshal(input)

	mockRepo.CreateCandidateCommentFunc = func(ctx context.Context, arg repository.CreateCandidateCommentParams) (repository.CandidateComment, error) {
		return repository.CandidateComment{
			ID:          pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
			CandidateID: arg.CandidateID,
			AuthorID:    arg.AuthorID,
			Content:     arg.Content,
			CreatedAt:   pgtype.Timestamptz{Time: time.Now(), Valid: true},
		}, nil
	}

	mockRepo.GetEmployeeFunc = func(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
		userID := pgtype.UUID{Bytes: [16]byte{9}, Valid: true}
		return repository.Employee{
			ID:        id,
			FirstName: "John",
			LastName:  "Doe",
			UserID:    userID,
		}, nil
	}
	mockRepo.GetUserByIDFunc = func(ctx context.Context, id pgtype.UUID) (repository.User, error) {
		return repository.User{
			ID:     id,
			Avatar: pgtype.Text{String: "avatar.png", Valid: true},
		}, nil
	}
	mockRepo.CheckIsHRFunc = func(ctx context.Context, id pgtype.UUID) (bool, error) {
		return true, nil
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/candidates/"+candidateID+"/comments", bytes.NewBuffer(body))
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}
