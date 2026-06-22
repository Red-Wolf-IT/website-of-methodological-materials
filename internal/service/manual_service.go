package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"website-of-methodological-materials/internal/models"
)

var ErrManualNotFound = errors.New("manual not found")

type ManualRepository interface {
	Create(ctx context.Context, input models.ManualCreate) (*models.Manual, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Manual, error)
}

type ManualService struct {
	repo ManualRepository
}

func NewManualService(repo ManualRepository) *ManualService {
	return &ManualService{repo: repo}
}

func (s *ManualService) Create(ctx context.Context, req models.CreateManualRequest) (*models.Manual, error) {
	manual, err := s.repo.Create(ctx, models.ManualCreate{
		Title:    req.Title,
		Author:   req.Author,
		Content:  req.Content,
		FilePath: req.FilePath,
	})
	if err != nil {
		return nil, fmt.Errorf("create manual: %w", err)
	}

	return manual, nil
}

func (s *ManualService) GetByID(ctx context.Context, id uuid.UUID) (*models.Manual, error) {
	manual, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrManualNotFound
		}
		return nil, fmt.Errorf("get manual: %w", err)
	}

	return manual, nil
}
