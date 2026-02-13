package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"hr-backend/internal/repository"
	"hr-backend/internal/utils"
	"hr-backend/test/mocks"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestListComments(t *testing.T) {
	mockRepo := &mocks.MockQuerier{}
	s := NewCandidateCommentService(mockRepo)

	candidateID := "00000000-0000-0000-0000-000000000001"
	candUUID, _ := utils.StringToUUID(candidateID)

	mockRepo.ListCandidateCommentsFunc = func(ctx context.Context, id pgtype.UUID) ([]repository.ListCandidateCommentsRow, error) {
		return []repository.ListCandidateCommentsRow{
			{
				ID:           pgtype.UUID{Bytes: [16]byte{1}, Valid: true},
				CandidateID:  candUUID,
				AuthorID:     pgtype.UUID{Bytes: [16]byte{2}, Valid: true},
				Content:      "Test comment",
				CreatedAt:    pgtype.Timestamptz{Time: time.Now(), Valid: true},
				AuthorName:   "John Doe",
				AuthorAvatar: pgtype.Text{String: "avatar.png", Valid: true},
				AuthorRole:   "HR",
			},
		}, nil
	}

	comments, err := s.ListComments(context.Background(), candidateID)
	assert.NoError(t, err)
	assert.Len(t, comments, 1)
	assert.Equal(t, "Test comment", comments[0].Content)
	assert.Equal(t, "John Doe", comments[0].AuthorName)
	assert.Equal(t, "HR", comments[0].AuthorRole)
}

func TestCreateComment(t *testing.T) {
	mockRepo := &mocks.MockQuerier{}
	s := NewCandidateCommentService(mockRepo)

	candidateID := "00000000-0000-0000-0000-000000000001"
	employeeID := "00000000-0000-0000-0000-000000000002"
	empUUID, _ := utils.StringToUUID(employeeID)
	content := "New comment"

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
		assert.Equal(t, empUUID, id)
		userID, _ := utils.StringToUUID("00000000-0000-0000-0000-000000000010")
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
	mockRepo.CheckRecruiterRoleFunc = func(ctx context.Context, id pgtype.UUID) (pgtype.UUID, error) {
		return pgtype.UUID{}, errors.New("not recruiter")
	}

	comment, err := s.CreateComment(context.Background(), candidateID, employeeID, content)
	assert.NoError(t, err)
	assert.NotNil(t, comment)
	assert.Equal(t, content, comment.Content)
	assert.Equal(t, "John Doe", comment.AuthorName)
	assert.Equal(t, "HR", comment.AuthorRole)
}

func TestDeleteComment(t *testing.T) {
	mockRepo := &mocks.MockQuerier{}
	s := NewCandidateCommentService(mockRepo)

	commentID := "00000000-0000-0000-0000-000000000001"
	userID := "00000000-0000-0000-0000-000000000003"
	employeeID := "00000000-0000-0000-0000-000000000002"
	empUUID, _ := utils.StringToUUID(employeeID)

	t.Run("Author can delete", func(t *testing.T) {
		mockRepo.GetCandidateCommentFunc = func(ctx context.Context, id pgtype.UUID) (repository.CandidateComment, error) {
			return repository.CandidateComment{
				ID:       id,
				AuthorID: empUUID,
			}, nil
		}
		mockRepo.DeleteCandidateCommentFunc = func(ctx context.Context, id pgtype.UUID) error {
			return nil
		}

		err := s.DeleteComment(context.Background(), commentID, userID, employeeID)
		assert.NoError(t, err)
	})

	t.Run("HR can delete", func(t *testing.T) {
		mockRepo.GetCandidateCommentFunc = func(ctx context.Context, id pgtype.UUID) (repository.CandidateComment, error) {
			return repository.CandidateComment{
				ID:       id,
				AuthorID: pgtype.UUID{Bytes: [16]byte{255}, Valid: true}, // Different author
			}, nil
		}
		mockRepo.CheckIsHRFunc = func(ctx context.Context, id pgtype.UUID) (bool, error) {
			return true, nil
		}
		mockRepo.DeleteCandidateCommentFunc = func(ctx context.Context, id pgtype.UUID) error {
			return nil
		}

		err := s.DeleteComment(context.Background(), commentID, userID, employeeID)
		assert.NoError(t, err)
	})

	t.Run("Author can delete without employee ID in context", func(t *testing.T) {
		userUUID, _ := utils.StringToUUID(userID)
		mockRepo.GetCandidateCommentFunc = func(ctx context.Context, id pgtype.UUID) (repository.CandidateComment, error) {
			return repository.CandidateComment{
				ID:       id,
				AuthorID: empUUID,
			}, nil
		}
		mockRepo.GetEmployeeByUserIDFunc = func(ctx context.Context, id pgtype.UUID) (repository.Employee, error) {
			assert.Equal(t, userUUID, id)
			return repository.Employee{ID: empUUID, UserID: userUUID}, nil
		}
		mockRepo.DeleteCandidateCommentFunc = func(ctx context.Context, id pgtype.UUID) error {
			return nil
		}

		err := s.DeleteComment(context.Background(), commentID, userID, "")
		assert.NoError(t, err)
	})

	t.Run("Others cannot delete", func(t *testing.T) {
		mockRepo.GetCandidateCommentFunc = func(ctx context.Context, id pgtype.UUID) (repository.CandidateComment, error) {
			return repository.CandidateComment{
				ID:       id,
				AuthorID: pgtype.UUID{Bytes: [16]byte{255}, Valid: true}, // Different author
			}, nil
		}
		mockRepo.CheckIsHRFunc = func(ctx context.Context, id pgtype.UUID) (bool, error) {
			return false, nil
		}

		err := s.DeleteComment(context.Background(), commentID, userID, employeeID)
		assert.Error(t, err)
		assert.Equal(t, ErrDeleteCommentNoPerm, err)
	})

	t.Run("Admin without employee profile cannot delete", func(t *testing.T) {
		mockRepo.GetCandidateCommentFunc = func(ctx context.Context, id pgtype.UUID) (repository.CandidateComment, error) {
			return repository.CandidateComment{
				ID:       id,
				AuthorID: pgtype.UUID{Bytes: [16]byte{255}, Valid: true},
			}, nil
		}

		err := s.DeleteComment(context.Background(), commentID, userID, "")
		assert.Error(t, err)
		assert.Equal(t, ErrDeleteCommentNoPerm, err)
	})

	t.Run("User without employee profile cannot delete", func(t *testing.T) {
		mockRepo.GetCandidateCommentFunc = func(ctx context.Context, id pgtype.UUID) (repository.CandidateComment, error) {
			return repository.CandidateComment{
				ID:       id,
				AuthorID: pgtype.UUID{Bytes: [16]byte{255}, Valid: true},
			}, nil
		}
		mockRepo.CheckIsHRFunc = func(ctx context.Context, id pgtype.UUID) (bool, error) {
			return false, nil
		}

		err := s.DeleteComment(context.Background(), commentID, userID, "")
		assert.Error(t, err)
		assert.Equal(t, ErrDeleteCommentNoPerm, err)
	})

	t.Run("Comment not found", func(t *testing.T) {
		mockRepo.GetCandidateCommentFunc = func(ctx context.Context, id pgtype.UUID) (repository.CandidateComment, error) {
			return repository.CandidateComment{}, errors.New("no rows")
		}

		err := s.DeleteComment(context.Background(), commentID, userID, employeeID)
		assert.Error(t, err)
	})
}
