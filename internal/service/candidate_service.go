package service

import (
	"context"
	"errors"
	"log"
	"strings"

	"hr-backend/internal/model"
	"hr-backend/internal/repository"
	"hr-backend/internal/utils"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

type CandidateService struct {
	repo                  repository.Querier
	txBeginner            TxBeginner
	notificationPublisher *NotificationPublisher
}

var (
	ErrReviewPermissionDenied  = errors.New("only assigned reviewer can submit review")
	ErrReviewerProfileNotFound = errors.New("reviewer profile not found")
	ErrCandidateNotFound       = errors.New("candidate not found")
)

func NewCandidateService(repo repository.Querier, txBeginner ...TxBeginner) *CandidateService {
	var beginner TxBeginner
	if len(txBeginner) > 0 {
		beginner = txBeginner[0]
	}

	return &CandidateService{
		repo:                  repo,
		txBeginner:            beginner,
		notificationPublisher: NewNotificationPublisher(repo),
	}
}

func (s *CandidateService) CreateCandidate(ctx context.Context, input model.CandidateInput) (*model.Candidate, error) {
	jobUUID, err := utils.StringToUUID(input.AppliedJobID)
	if err != nil {
		return nil, err
	}

	status := input.Status
	if status == "" {
		status = "new"
	}

	params := repository.CreateCandidateParams{
		Name:            input.Name,
		Avatar:          pgtype.Text{String: input.Avatar, Valid: input.Avatar != ""},
		Email:           input.Email,
		Phone:           input.Phone,
		ExperienceYears: int32(input.ExperienceYears),
		Education:       input.Education,
		AppliedJobID:    jobUUID,
		Channel:         input.Channel,
		ResumeUrl:       input.ResumeURL,
		Status:          status,
		AppliedAt:       pgtype.Timestamptz{Time: input.AppliedAt, Valid: true},
	}

	candidate, err := s.repo.CreateCandidate(ctx, params)
	if err != nil {
		return nil, err
	}

	return s.GetCandidate(ctx, utils.UUIDToString(candidate.ID))
}

func (s *CandidateService) ListCandidates(ctx context.Context, jobIDFilter string, reviewerIDFilter string, reviewStatusFilter string, statusFilter string, search string, page int, limit int) ([]model.Candidate, int64, error) {
	params := repository.ListCandidatesParams{
		JobID:        pgtype.UUID{Valid: false},
		ReviewerID:   pgtype.UUID{Valid: false},
		ReviewStatus: pgtype.Text{Valid: false},
		Status:       pgtype.Text{Valid: false},
		Search:       pgtype.Text{Valid: false},
		Limit:        int32(limit),
		Offset:       int32((page - 1) * limit),
	}

	countParams := repository.CountCandidatesParams{
		JobID:        pgtype.UUID{Valid: false},
		ReviewerID:   pgtype.UUID{Valid: false},
		ReviewStatus: pgtype.Text{Valid: false},
		Status:       pgtype.Text{Valid: false},
		Search:       pgtype.Text{Valid: false},
	}

	if jobIDFilter != "" && jobIDFilter != "all" {
		uuid, err := utils.StringToUUID(jobIDFilter)
		if err != nil {
			return nil, 0, err
		}
		params.JobID = uuid
		countParams.JobID = uuid
	}

	if reviewerIDFilter != "" {
		uuid, err := utils.StringToUUID(reviewerIDFilter)
		if err != nil {
			return nil, 0, err
		}
		params.ReviewerID = uuid
		countParams.ReviewerID = uuid
	}

	if reviewStatusFilter != "" {
		params.ReviewStatus = pgtype.Text{String: reviewStatusFilter, Valid: true}
		countParams.ReviewStatus = pgtype.Text{String: reviewStatusFilter, Valid: true}
	}

	if statusFilter != "" {
		params.Status = pgtype.Text{String: statusFilter, Valid: true}
		countParams.Status = pgtype.Text{String: statusFilter, Valid: true}
	}

	if search != "" {
		params.Search = pgtype.Text{String: search, Valid: true}
		countParams.Search = pgtype.Text{String: search, Valid: true}
	}

	// efficient parallel fetch if possible, but sequential is fine for now
	rows, err := s.repo.ListCandidates(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountCandidates(ctx, countParams)
	if err != nil {
		return nil, 0, err
	}

	result := make([]model.Candidate, len(rows))
	for i, r := range rows {
		result[i] = mapCandidateRowToModel(r)
	}
	return result, total, nil
}

func (s *CandidateService) GetCandidateCountsByJob(ctx context.Context) (map[string]int64, error) {
	rows, err := s.repo.GetCandidateCountsByJob(ctx)
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, row := range rows {
		if row.AppliedJobID.Valid {
			counts[utils.UUIDToString(row.AppliedJobID)] = row.Count
		}
	}
	return counts, nil
}

func (s *CandidateService) AssignReviewer(ctx context.Context, id string, reviewerID string, assignedByUserID string) (*model.Candidate, error) {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	reviewerUUID, err := utils.StringToUUID(reviewerID)
	if err != nil {
		return nil, err
	}

	assignedByUserUUID, err := utils.StringToUUID(assignedByUserID)
	if err != nil {
		return nil, err
	}

	var row repository.AssignReviewerRow
	assignCore := func(q repository.Querier) error {
		assignedRow, assignErr := q.AssignReviewer(ctx, repository.AssignReviewerParams{
			ID:         uuid,
			ReviewerID: reviewerUUID,
		})
		if assignErr != nil {
			return assignErr
		}

		if removeErr := q.UpdateCandidateReviewerRemovedAt(ctx, uuid); removeErr != nil {
			return removeErr
		}

		if _, insertErr := q.InsertCandidateReviewer(ctx, repository.InsertCandidateReviewerParams{
			CandidateID:      uuid,
			ReviewerID:       reviewerUUID,
			AssignedByUserID: assignedByUserUUID,
		}); insertErr != nil {
			return insertErr
		}

		row = assignedRow
		return nil
	}

	if s.txBeginner != nil {
		err = runInTx(ctx, s.txBeginner, func(txQueries *repository.Queries) error {
			return assignCore(txQueries)
		})
	} else {
		err = assignCore(s.repo)
	}
	if err != nil {
		return nil, err
	}

	if err := s.notificationPublisher.PublishCandidateReviewerAssigned(ctx, reviewerUUID, uuid); err != nil {
		log.Printf(
			"warn: failed to publish notification event=%s reviewer_id=%s candidate_id=%s err=%v",
			model.NotificationEventCandidateReviewerAssigned,
			utils.UUIDToString(reviewerUUID),
			id,
			err,
		)
	}

	return mapAssignReviewerRowToModel(row), nil
}

func (s *CandidateService) SubmitReview(
	ctx context.Context,
	id string,
	userID string,
	status string,
	comment string,
) (*model.Candidate, error) {
	candidateID, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	userUUID, err := utils.StringToUUID(userID)
	if err != nil {
		return nil, err
	}

	var row repository.SubmitReviewRow
	shouldNotifyRecruiter := false
	recruiterUserID := pgtype.UUID{}
	reviewerName := ""

	submitCore := func(q repository.Querier) error {
		employee, getEmployeeErr := q.GetEmployeeByUserID(ctx, userUUID)
		if getEmployeeErr != nil {
			if errors.Is(getEmployeeErr, pgx.ErrNoRows) {
				return ErrReviewerProfileNotFound
			}
			return getEmployeeErr
		}

		candidate, getCandidateErr := q.GetCandidate(ctx, candidateID)
		if getCandidateErr != nil {
			if errors.Is(getCandidateErr, pgx.ErrNoRows) {
				return ErrCandidateNotFound
			}
			return getCandidateErr
		}

		assignment, getAssignmentErr := q.GetReviewerAssignment(ctx, repository.GetReviewerAssignmentParams{
			CandidateID: candidateID,
			ReviewerID:  employee.ID,
		})
		assignmentExists := true
		if getAssignmentErr != nil {
			// Legacy fallback for rows created before assignment-table based review status.
			if errors.Is(getAssignmentErr, pgx.ErrNoRows) {
				assignmentExists = false
				if !candidate.ReviewerID.Valid || utils.UUIDToString(candidate.ReviewerID) != utils.UUIDToString(employee.ID) {
					return ErrReviewPermissionDenied
				}
			} else {
				return getAssignmentErr
			}
		}
		if !assignmentExists &&
			(!candidate.ReviewerID.Valid || utils.UUIDToString(candidate.ReviewerID) != utils.UUIDToString(employee.ID)) {
			return ErrReviewPermissionDenied
		}

		submittedRow, submitErr := q.SubmitReview(ctx, repository.SubmitReviewParams{
			ID:           candidateID,
			ReviewStatus: pgtype.Text{String: status, Valid: true},
		})
		if submitErr != nil {
			return submitErr
		}

		if updateErr := q.UpdateCandidateReviewerReviewStatus(ctx, repository.UpdateCandidateReviewerReviewStatusParams{
			CandidateID:  candidateID,
			ReviewerID:   employee.ID,
			ReviewStatus: status,
		}); updateErr != nil {
			return updateErr
		}

		trimmedComment := strings.TrimSpace(comment)
		if trimmedComment != "" {
			if _, createCommentErr := q.CreateCandidateComment(ctx, repository.CreateCandidateCommentParams{
				CandidateID: candidateID,
				AuthorID:    employee.ID,
				Content:     trimmedComment,
				CommentType: "normal",
			}); createCommentErr != nil {
				return createCommentErr
			}
		}

		if decisionCommentType, ok := reviewStatusToDecisionCommentType(status); ok {
			if _, createDecisionErr := q.CreateCandidateComment(ctx, repository.CreateCandidateCommentParams{
				CandidateID: candidateID,
				AuthorID:    employee.ID,
				Content:     status,
				CommentType: decisionCommentType,
			}); createDecisionErr != nil {
				return createDecisionErr
			}
		}

		if deleteErr := q.DeleteNotificationsBySubjectAndType(ctx, repository.DeleteNotificationsBySubjectAndTypeParams{
			UserID:      userUUID,
			SubjectType: model.NotificationSubjectTypeCandidate,
			SubjectID:   candidateID,
			EventType:   model.NotificationEventCandidateReviewerAssigned,
		}); deleteErr != nil {
			log.Printf(
				"warn: failed to delete notification event=%s reviewer_user_id=%s candidate_id=%s err=%v",
				model.NotificationEventCandidateReviewerAssigned,
				userID,
				id,
				deleteErr,
			)
		}

		if status != "pending" && assignmentExists && assignment.AssignedByUserID.Valid {
			shouldNotifyRecruiter = true
			recruiterUserID = assignment.AssignedByUserID
			reviewerName = strings.TrimSpace(employee.FirstName + " " + employee.LastName)
		}

		row = submittedRow
		return nil
	}

	if s.txBeginner != nil {
		err = runInTx(ctx, s.txBeginner, func(txQueries *repository.Queries) error {
			return submitCore(txQueries)
		})
	} else {
		err = submitCore(s.repo)
	}
	if err != nil {
		return nil, err
	}

	if shouldNotifyRecruiter {
		if notifyErr := s.notificationPublisher.PublishReviewCompleted(
			ctx,
			recruiterUserID,
			candidateID,
			status,
			reviewerName,
		); notifyErr != nil {
			log.Printf(
				"warn: failed to publish notification event=%s recruiter_user_id=%s candidate_id=%s err=%v",
				model.NotificationEventReviewCompleted,
				utils.UUIDToString(recruiterUserID),
				id,
				notifyErr,
			)
		}
	}

	return mapSubmitReviewRowToModel(row), nil
}

func reviewStatusToDecisionCommentType(status string) (string, bool) {
	switch status {
	case "suitable":
		return "review_suitable", true
	case "unsuitable":
		return "review_unsuitable", true
	default:
		return "", false
	}
}

func (s *CandidateService) GetCandidate(ctx context.Context, id string) (*model.Candidate, error) {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	row, err := s.repo.GetCandidate(ctx, uuid)
	if err != nil {
		return nil, err
	}

	return &model.Candidate{
		ID:              utils.UUIDToString(row.ID),
		Name:            row.Name,
		Avatar:          row.Avatar.String,
		Email:           row.Email,
		Phone:           row.Phone,
		ExperienceYears: int(row.ExperienceYears),
		Education:       row.Education,
		AppliedJobID:    utils.UUIDToString(row.AppliedJobID),
		AppliedJobTitle: row.AppliedJobTitle,
		Channel:         row.Channel,
		ResumeURL:       row.ResumeUrl,
		Status:          row.Status,
		AppliedAt:       row.AppliedAt.Time,
		ReviewerID:      utils.UUIDToString(row.ReviewerID),
		ReviewStatus:    row.ReviewStatus.String,
	}, nil
}

func (s *CandidateService) UpdateCandidate(ctx context.Context, id string, input model.CandidateInput) (*model.Candidate, error) {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	jobUUID, err := utils.StringToUUID(input.AppliedJobID)
	if err != nil {
		return nil, err
	}

	params := repository.UpdateCandidateParams{
		ID:              uuid,
		Name:            input.Name,
		Avatar:          pgtype.Text{String: input.Avatar, Valid: input.Avatar != ""},
		Email:           input.Email,
		Phone:           input.Phone,
		ExperienceYears: int32(input.ExperienceYears),
		Education:       input.Education,
		AppliedJobID:    jobUUID,
		Channel:         input.Channel,
		ResumeUrl:       input.ResumeURL,
		Status:          input.Status,
		AppliedAt:       pgtype.Timestamptz{Time: input.AppliedAt, Valid: true},
	}

	_, err = s.repo.UpdateCandidate(ctx, params)
	if err != nil {
		return nil, err
	}

	return s.GetCandidate(ctx, id)
}

func (s *CandidateService) UpdateStatus(ctx context.Context, id string, status string) (*model.Candidate, error) {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.UpdateCandidateStatus(ctx, repository.UpdateCandidateStatusParams{
		ID:     uuid,
		Status: status,
	})
	if err != nil {
		return nil, err
	}

	if status == "hired" || status == "rejected" {
		_ = s.repo.UpdateCandidateReviewerRemovedAt(ctx, uuid)
	}

	return s.GetCandidate(ctx, id)
}

func (s *CandidateService) UpdateResume(ctx context.Context, id string, resumeUrl string) (*model.Candidate, error) {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return nil, err
	}

	_, err = s.repo.UpdateCandidateResume(ctx, repository.UpdateCandidateResumeParams{
		ID:        uuid,
		ResumeUrl: resumeUrl,
	})
	if err != nil {
		return nil, err
	}
	return s.GetCandidate(ctx, id)
}

func (s *CandidateService) DeleteCandidate(ctx context.Context, id string) error {
	uuid, err := utils.StringToUUID(id)
	if err != nil {
		return err
	}
	return s.repo.DeleteCandidate(ctx, uuid)
}

func mapCandidateRowToModel(row repository.ListCandidatesRow) model.Candidate {
	return model.Candidate{
		ID:              utils.UUIDToString(row.ID),
		Name:            row.Name,
		Avatar:          row.Avatar.String,
		Email:           row.Email,
		Phone:           row.Phone,
		ExperienceYears: int(row.ExperienceYears),
		Education:       row.Education,
		AppliedJobID:    utils.UUIDToString(row.AppliedJobID),
		AppliedJobTitle: row.AppliedJobTitle,
		Channel:         row.Channel,
		ResumeURL:       row.ResumeUrl,
		Status:          row.Status,
		AppliedAt:       row.AppliedAt.Time,
		ReviewerID:      utils.UUIDToString(row.ReviewerID),
		ReviewStatus:    row.ReviewStatus.String,
	}
}

func mapAssignReviewerRowToModel(row repository.AssignReviewerRow) *model.Candidate {
	return &model.Candidate{
		ID:              utils.UUIDToString(row.ID),
		Name:            row.Name,
		Avatar:          row.Avatar.String,
		Email:           row.Email,
		Phone:           row.Phone,
		ExperienceYears: int(row.ExperienceYears),
		Education:       row.Education,
		AppliedJobID:    utils.UUIDToString(row.AppliedJobID),
		AppliedJobTitle: row.AppliedJobTitle,
		Channel:         row.Channel,
		ResumeURL:       row.ResumeUrl,
		Status:          row.Status,
		AppliedAt:       row.AppliedAt.Time,
		ReviewerID:      utils.UUIDToString(row.ReviewerID),
		ReviewStatus:    row.ReviewStatus.String,
	}
}

func mapSubmitReviewRowToModel(row repository.SubmitReviewRow) *model.Candidate {
	return &model.Candidate{
		ID:              utils.UUIDToString(row.ID),
		Name:            row.Name,
		Avatar:          row.Avatar.String,
		Email:           row.Email,
		Phone:           row.Phone,
		ExperienceYears: int(row.ExperienceYears),
		Education:       row.Education,
		AppliedJobID:    utils.UUIDToString(row.AppliedJobID),
		AppliedJobTitle: row.AppliedJobTitle,
		Channel:         row.Channel,
		ResumeURL:       row.ResumeUrl,
		Status:          row.Status,
		AppliedAt:       row.AppliedAt.Time,
		ReviewerID:      utils.UUIDToString(row.ReviewerID),
		ReviewStatus:    row.ReviewStatus.String,
	}
}
