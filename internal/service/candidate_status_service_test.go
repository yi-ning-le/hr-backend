package service

import (
	"context"
	"testing"

	"hr-backend/internal/repository"
	"hr-backend/test/mocks"

	"github.com/stretchr/testify/assert"
)

func TestReorderStatuses_ReturnsErrorOnInvalidID(t *testing.T) {
	updateCalled := false
	mockRepo := &mocks.MockQuerier{
		UpdateCandidateStatusOrderFunc: func(ctx context.Context, arg repository.UpdateCandidateStatusOrderParams) error {
			updateCalled = true
			return nil
		},
	}
	svc := NewCandidateStatusService(mockRepo)

	err := svc.ReorderStatuses(context.Background(), []string{
		"00000000-0000-0000-0000-000000000001",
		"invalid-uuid",
	})

	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidStatusID)
	assert.False(t, updateCalled)
}
