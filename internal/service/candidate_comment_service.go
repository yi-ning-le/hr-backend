package service

import (
	"context"
	"errors"
	"fmt"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/utils"

	"github.com/jackc/pgx/v5"
)

type CandidateCommentService struct {
	repo repository.Querier
}

var (
	ErrCommentNotFound     = errors.New("comment not found")
	ErrDeleteCommentNoPerm = errors.New("you do not have permission to delete this comment")
)

func NewCandidateCommentService(repo repository.Querier) *CandidateCommentService {
	return &CandidateCommentService{repo: repo}
}

func (s *CandidateCommentService) ListComments(ctx context.Context, candidateID string) ([]model.CandidateComment, error) {
	uuid, err := utils.StringToUUID(candidateID)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.ListCandidateComments(ctx, uuid)
	if err != nil {
		return nil, err
	}

	comments := make([]model.CandidateComment, len(rows))
	for i, row := range rows {
		authorName := ""
		if row.AuthorName != nil {
			authorName = fmt.Sprintf("%v", row.AuthorName)
		}

		comments[i] = model.CandidateComment{
			ID:           utils.UUIDToString(row.ID),
			CandidateID:  utils.UUIDToString(row.CandidateID),
			AuthorID:     utils.UUIDToString(row.AuthorID),
			AuthorName:   authorName,
			AuthorAvatar: row.AuthorAvatar.String,
			AuthorRole:   row.AuthorRole,
			Content:      row.Content,
			CreatedAt:    row.CreatedAt.Time,
		}
	}

	return comments, nil
}

func (s *CandidateCommentService) CreateComment(ctx context.Context, candidateID string, employeeID string, content string) (*model.CandidateComment, error) {
	candUUID, err := utils.StringToUUID(candidateID)
	if err != nil {
		return nil, err
	}

	empUUID, err := utils.StringToUUID(employeeID)
	if err != nil {
		return nil, err
	}

	params := repository.CreateCandidateCommentParams{
		CandidateID: candUUID,
		AuthorID:    empUUID,
		Content:     content,
	}

	created, err := s.repo.CreateCandidateComment(ctx, params)
	if err != nil {
		return nil, err
	}

	author, err := s.repo.GetEmployee(ctx, created.AuthorID)
	if err != nil {
		return nil, err
	}

	user, err := s.repo.GetUserByID(ctx, author.UserID)
	if err != nil {
		return nil, err
	}

	authorRole := "INTERVIEWER"
	if isHR, checkErr := s.repo.CheckIsHR(ctx, created.AuthorID); checkErr == nil && isHR {
		authorRole = "HR"
	} else if recruiterID, checkErr := s.repo.CheckRecruiterRole(ctx, created.AuthorID); checkErr == nil && recruiterID.Valid {
		authorRole = "HR"
	}

	return &model.CandidateComment{
		ID:           utils.UUIDToString(created.ID),
		CandidateID:  utils.UUIDToString(created.CandidateID),
		AuthorID:     utils.UUIDToString(created.AuthorID),
		AuthorName:   fmt.Sprintf("%s %s", author.FirstName, author.LastName),
		AuthorAvatar: user.Avatar.String,
		AuthorRole:   authorRole,
		Content:      created.Content,
		CreatedAt:    created.CreatedAt.Time,
	}, nil
}

func (s *CandidateCommentService) DeleteComment(ctx context.Context, commentID string, userID string, employeeID string) error {
	commUUID, err := utils.StringToUUID(commentID)
	if err != nil {
		return err
	}

	empUUID, err := utils.StringToUUID(employeeID)
	if err != nil {
		return err
	}

	userUUID, err := utils.StringToUUID(userID)
	if err != nil {
		return err
	}

	comment, err := s.repo.GetCandidateComment(ctx, commUUID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrCommentNotFound
		}
		return err
	}

	// Permission check: Author OR HR OR Admin
	isAuthor := utils.UUIDToString(comment.AuthorID) == utils.UUIDToString(empUUID)
	if isAuthor {
		return s.repo.DeleteCandidateComment(ctx, commUUID)
	}

	// Check Admin
	isAdmin, err := s.repo.CheckIsAdmin(ctx, userUUID)
	if err == nil && isAdmin {
		return s.repo.DeleteCandidateComment(ctx, commUUID)
	}

	// Check HR
	isHR, err := s.repo.CheckIsHR(ctx, empUUID)
	if err == nil && isHR {
		return s.repo.DeleteCandidateComment(ctx, commUUID)
	}

	return ErrDeleteCommentNoPerm
}
