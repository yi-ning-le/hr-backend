package handler

import (
	"hr-backend/internal/repository"
	"hr-backend/internal/service"
)

type RecruitmentHandler struct {
	queries            repository.Querier
	recruitmentService *service.RecruitmentService
}

func NewRecruitmentHandler(q repository.Querier) *RecruitmentHandler {
	return &RecruitmentHandler{
		queries:            q,
		recruitmentService: service.NewRecruitmentService(q),
	}
}
