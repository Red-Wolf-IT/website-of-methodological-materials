package service

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"website-of-methodological-materials/internal/models"
	"website-of-methodological-materials/internal/repository/postgres"
)

var (
	ErrManualNotFound    = errors.New("manual not found")
	ErrTagNotFound       = errors.New("tag not found")
	ErrInvalidListParams = errors.New("invalid list params")
	ErrFileNotFound      = errors.New("file not found")
)

type FileStorage interface {
	Save(manualID uuid.UUID, originalName string, src io.Reader) (string, error)
	Open(webPath string) (io.ReadCloser, error)
	Remove(webPath string) error
}

const (
	defaultListPage  = 1
	defaultListLimit = 20
	maxListLimit     = 100
)

type ManualRepository interface {
	Create(ctx context.Context, input models.ManualCreate) (*models.Manual, error)
	GetByID(ctx context.Context, id uuid.UUID) (*models.Manual, error)
	List(ctx context.Context, filter models.ManualListFilter) (*models.ManualListResult, error)
	Update(ctx context.Context, id uuid.UUID, input models.ManualUpdate) (*models.Manual, error)
	UpdateFilePath(ctx context.Context, id uuid.UUID, filePath string) (*models.Manual, error)
	Delete(ctx context.Context, id uuid.UUID) (*string, error)
	AttachTags(ctx context.Context, manualID uuid.UUID, tagIDs []int) error
}

type ManualService struct {
	repo  ManualRepository
	files FileStorage
}

func NewManualService(repo ManualRepository, files FileStorage) *ManualService {
	return &ManualService{repo: repo, files: files}
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

func (s *ManualService) List(ctx context.Context, filter models.ManualListFilter) (*models.ManualListResult, error) {
	normalized, err := normalizeListFilter(filter)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrInvalidListParams, err)
	}

	result, err := s.repo.List(ctx, normalized)
	if err != nil {
		return nil, fmt.Errorf("list manuals: %w", err)
	}

	return result, nil
}

func normalizeListFilter(filter models.ManualListFilter) (models.ManualListFilter, error) {
	if filter.Page <= 0 {
		filter.Page = defaultListPage
	}
	if filter.Limit <= 0 {
		filter.Limit = defaultListLimit
	}
	if filter.Limit > maxListLimit {
		return filter, fmt.Errorf("limit must be at most %d", maxListLimit)
	}

	switch filter.Sort {
	case "", "popular":
	default:
		return filter, fmt.Errorf("sort must be empty or popular")
	}

	return filter, nil
}

func (s *ManualService) Update(ctx context.Context, id uuid.UUID, req models.UpdateManualRequest) (*models.Manual, error) {
	filePath := req.FilePath
	if filePath == nil {
		existing, err := s.repo.GetByID(ctx, id)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, ErrManualNotFound
			}
			return nil, fmt.Errorf("get manual for update: %w", err)
		}
		filePath = existing.FilePath
	}

	manual, err := s.repo.Update(ctx, id, models.ManualUpdate{
		Title:    req.Title,
		Author:   req.Author,
		Content:  req.Content,
		FilePath: filePath,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrManualNotFound
		}
		return nil, fmt.Errorf("update manual: %w", err)
	}

	return manual, nil
}

func (s *ManualService) Delete(ctx context.Context, id uuid.UUID) error {
	filePath, err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrManualNotFound
		}
		return fmt.Errorf("delete manual: %w", err)
	}

	if filePath != nil && s.files != nil {
		_ = s.files.Remove(*filePath)
	}

	return nil
}

func (s *ManualService) UploadAttachment(ctx context.Context, id uuid.UUID, filename string, src io.Reader) (*models.Manual, error) {
	manual, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrManualNotFound
		}
		return nil, fmt.Errorf("get manual: %w", err)
	}

	if manual.FilePath != nil && s.files != nil {
		_ = s.files.Remove(*manual.FilePath)
	}

	webPath, err := s.files.Save(id, filename, src)
	if err != nil {
		return nil, fmt.Errorf("save file: %w", err)
	}

	updated, err := s.repo.UpdateFilePath(ctx, id, webPath)
	if err != nil {
		_ = s.files.Remove(webPath)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrManualNotFound
		}
		return nil, fmt.Errorf("update file path: %w", err)
	}

	return updated, nil
}

func (s *ManualService) OpenAttachment(webPath string) (io.ReadCloser, error) {
	if s.files == nil {
		return nil, ErrFileNotFound
	}

	file, err := s.files.Open(webPath)
	if err != nil {
		return nil, ErrFileNotFound
	}

	return file, nil
}

func (s *ManualService) AttachTags(ctx context.Context, manualID uuid.UUID, tagIDs []int) (*models.Manual, error) {
	if err := s.repo.AttachTags(ctx, manualID, tagIDs); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrManualNotFound
		}

		var tagErr *postgres.TagNotFoundError
		if errors.As(err, &tagErr) {
			return nil, ErrTagNotFound
		}

		return nil, fmt.Errorf("attach tags: %w", err)
	}

	manual, err := s.repo.GetByID(ctx, manualID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrManualNotFound
		}
		return nil, fmt.Errorf("get manual after attach: %w", err)
	}

	return manual, nil
}
