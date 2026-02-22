package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"hr-backend/internal/repository"
	"hr-backend/internal/service"
	"hr-backend/test/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestCandidateStatusHandler_ReorderStatuses_InvalidIDReturns400(t *testing.T) {
	gin.SetMode(gin.TestMode)

	mockRepo := &mocks.MockQuerier{
		UpdateCandidateStatusOrderFunc: func(ctx context.Context, arg repository.UpdateCandidateStatusOrderParams) error {
			return nil
		},
	}
	svc := service.NewCandidateStatusService(mockRepo)
	h := NewCandidateStatusHandler(svc)

	r := gin.New()
	r.PATCH("/candidate-statuses/reorder", h.ReorderStatuses)

	body, _ := json.Marshal(map[string][]string{
		"ids": {"invalid-uuid"},
	})
	req, _ := http.NewRequest(http.MethodPatch, "/candidate-statuses/reorder", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var resp map[string]string
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["error"], "invalid status id")
}
