package service

import (
	"sorint-fleet/internal/dto"
	"sorint-fleet/internal/model"
	"sorint-fleet/internal/repository"
	"sorint-fleet/internal/search"
	"time"

	"github.com/google/uuid"
)

type RequestService struct {
	repo *repository.RequestRepository
}

func NewRequestService(repo *repository.RequestRepository) *RequestService {
	return &RequestService{repo: repo}
}

func (s *RequestService) Create(dto dto.CreateRequestDto) (*model.Request, error) {
	var uidPtr *uuid.UUID
	if dto.UserID != "" {
		parsed, err := uuid.Parse(dto.UserID)
		if err != nil {
			return nil, err
		}
		uidPtr = &parsed
	}
	r := &model.Request{
		UserID:    uidPtr,
		QueryText: dto.QueryText,
		CreatedAt: time.Now(),
	}
	if err := s.repo.Create(r); err != nil {
		return nil, err
	}
	_ = search.IndexRequest(*r)
	return r, nil
}

func (s *RequestService) Search(q string, size int) ([]map[string]interface{}, error) {
	return search.SearchRequests(q, size)
}
