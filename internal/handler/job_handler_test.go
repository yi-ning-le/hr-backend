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
	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/service"
	"hr-backend/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestListJobsHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		ListJobsFunc: func(ctx context.Context) ([]repository.Job, error) {
			return []repository.Job{
				{
					ID:    pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
					Title: "Job 1",
				},
			}, nil
		},
	}

	svc := service.NewJobService(mockRepo)
	h := handler.NewJobHandler(svc)

	r := gin.New()
	r.GET("/jobs", h.ListJobs)

	req, _ := http.NewRequest("GET", "/jobs", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var jobs []model.JobPosition
	if err := json.Unmarshal(w.Body.Bytes(), &jobs); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if len(jobs) != 1 {
		t.Errorf("expected 1 job, got %d", len(jobs))
	}
}

func TestCreateJobHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		CreateJobFunc: func(ctx context.Context, arg repository.CreateJobParams) (repository.Job, error) {
			return repository.Job{
				ID:    pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				Title: arg.Title,
			}, nil
		},
	}

	svc := service.NewJobService(mockRepo)
	h := handler.NewJobHandler(svc)

	r := gin.New()
	r.POST("/jobs", h.CreateJob)

	input := model.JobInput{
		Title:          "New Job",
		Department:     "IT",
		HeadCount:      1,
		OpenDate:       time.Now(),
		JobDescription: "Desc",
	}
	body, _ := json.Marshal(input)
	req, _ := http.NewRequest("POST", "/jobs", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d, body: %s", w.Code, w.Body.String())
	}
}
