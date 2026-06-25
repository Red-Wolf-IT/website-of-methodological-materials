package service

import (
	"context"
	"errors"
	"fmt"

	"website-of-methodological-materials/internal/models"
	"website-of-methodological-materials/internal/repository/postgres"
)

var ErrTagNameTaken = errors.New("tag name already exists")

type TagRepository interface {
	Create(ctx context.Context, name string) (*models.Tag, error)
	List(ctx context.Context) ([]models.Tag, error)
}

type TagService struct {
	repo TagRepository
}

func NewTagService(repo TagRepository) *TagService {
	return &TagService{repo: repo}
}

func (s *TagService) Create(ctx context.Context, req models.CreateTagRequest) (*models.Tag, error) {
	tag, err := s.repo.Create(ctx, req.Name)
	if err != nil {
		if errors.Is(err, postgres.ErrTagNameTaken) {
			return nil, ErrTagNameTaken
		}
		return nil, fmt.Errorf("create tag: %w", err)
	}

	return tag, nil
}

func (s *TagService) List(ctx context.Context) ([]models.Tag, error) {
	tags, err := s.repo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("list tags: %w", err)
	}

	return tags, nil
}
