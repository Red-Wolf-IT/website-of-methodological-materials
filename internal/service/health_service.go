package service

import (
	"context"

	"website-of-methodological-materials/internal/models"
)

// HealthRepository — интерфейс, чтобы в тестах подменять реальную БД
type HealthRepository interface {
	Ping(ctx context.Context) error
}

type HealthService struct {
	repo HealthRepository
}

func NewHealthService(repo HealthRepository) *HealthService {
	return &HealthService{repo: repo}
}

func (s *HealthService) Check(ctx context.Context) models.HealthResponse {
	if err := s.repo.Ping(ctx); err != nil {
		return models.HealthResponse{
			Status:   "error",
			Database: "unavailable",
		}
	}

	return models.HealthResponse{
		Status:   "ok",
		Database: "ok",
	}
}
