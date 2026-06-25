package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"website-of-methodological-materials/internal/models"
)

type ManualRepository struct {
	pool *pgxpool.Pool
}

func NewManualRepository(pool *pgxpool.Pool) *ManualRepository {
	return &ManualRepository{pool: pool}
}

// Create вставляет запись и возвращает её с id/timestamps из RETURNING
func (r *ManualRepository) Create(ctx context.Context, input models.ManualCreate) (*models.Manual, error) {
	const query = `
		INSERT INTO manuals (title, author, content, file_path)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, author, content, file_path, views_count, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, input.Title, input.Author, input.Content, input.FilePath)

	manual, err := scanManual(row)
	if err != nil {
		return nil, fmt.Errorf("insert manual: %w", err)
	}

	return manual, nil
}

func (r *ManualRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Manual, error) {
	return r.GetByIDWithTags(ctx, id)
}

func (r *ManualRepository) GetByIDWithTags(ctx context.Context, id uuid.UUID) (*models.Manual, error) {
	const query = `
		SELECT
			m.id, m.title, m.author, m.content, m.file_path, m.views_count, m.created_at, m.updated_at,
			t.id, t.name
		FROM manuals m
		LEFT JOIN manual_tags mt ON mt.manual_id = m.id
		LEFT JOIN tags t ON t.id = mt.tag_id
		WHERE m.id = $1
		ORDER BY t.name NULLS LAST
	`

	rows, err := r.pool.Query(ctx, query, id)
	if err != nil {
		return nil, fmt.Errorf("select manual with tags: %w", err)
	}
	defer rows.Close()

	var manual *models.Manual
	tags := make([]models.Tag, 0)

	for rows.Next() {
		var row models.Manual
		var tagID *int
		var tagName *string

		if err := rows.Scan(
			&row.ID,
			&row.Title,
			&row.Author,
			&row.Content,
			&row.FilePath,
			&row.ViewsCount,
			&row.CreatedAt,
			&row.UpdatedAt,
			&tagID,
			&tagName,
		); err != nil {
			return nil, fmt.Errorf("scan manual with tags: %w", err)
		}

		if manual == nil {
			manual = &row
		}

		if tagID != nil && tagName != nil {
			tags = append(tags, models.Tag{ID: *tagID, Name: *tagName})
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate manual with tags: %w", err)
	}

	if manual == nil {
		return nil, pgx.ErrNoRows
	}

	manual.Tags = tags
	return manual, nil
}

func (r *ManualRepository) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM manuals WHERE id = $1)`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("check manual exists: %w", err)
	}

	return exists, nil
}

func (r *ManualRepository) AttachTags(ctx context.Context, manualID uuid.UUID, tagIDs []int) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var manualExists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM manuals WHERE id = $1)`, manualID).Scan(&manualExists); err != nil {
		return fmt.Errorf("check manual exists: %w", err)
	}
	if !manualExists {
		return pgx.ErrNoRows
	}

	const insertQuery = `
		INSERT INTO manual_tags (manual_id, tag_id)
		VALUES ($1, $2)
		ON CONFLICT DO NOTHING
	`

	for _, tagID := range tagIDs {
		var tagExists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM tags WHERE id = $1)`, tagID).Scan(&tagExists); err != nil {
			return fmt.Errorf("check tag exists: %w", err)
		}
		if !tagExists {
			return &TagNotFoundError{TagID: tagID}
		}

		if _, err := tx.Exec(ctx, insertQuery, manualID, tagID); err != nil {
			return fmt.Errorf("insert manual_tag: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

func (r *ManualRepository) IncrementViews(ctx context.Context, id uuid.UUID) error {
	const query = `
		UPDATE manuals
		SET views_count = views_count + 1
		WHERE id = $1
	`

	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("increment views: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}

	return nil
}

func (r *ManualRepository) Update(ctx context.Context, id uuid.UUID, input models.ManualUpdate) (*models.Manual, error) {
	const query = `
		UPDATE manuals
		SET title = $2, author = $3, content = $4, file_path = $5, updated_at = now()
		WHERE id = $1
		RETURNING id, title, author, content, file_path, views_count, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, id, input.Title, input.Author, input.Content, input.FilePath)

	manual, err := scanManual(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("update manual: %w", err)
	}

	return manual, nil
}

func (r *ManualRepository) UpdateFilePath(ctx context.Context, id uuid.UUID, filePath string) (*models.Manual, error) {
	const query = `
		UPDATE manuals
		SET file_path = $2, updated_at = now()
		WHERE id = $1
		RETURNING id, title, author, content, file_path, views_count, created_at, updated_at
	`

	row := r.pool.QueryRow(ctx, query, id, filePath)

	manual, err := scanManual(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("update file path: %w", err)
	}

	return manual, nil
}

func (r *ManualRepository) Delete(ctx context.Context, id uuid.UUID) (*string, error) {
	const query = `DELETE FROM manuals WHERE id = $1 RETURNING file_path`

	var filePath *string
	err := r.pool.QueryRow(ctx, query, id).Scan(&filePath)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("delete manual: %w", err)
	}

	return filePath, nil
}

type TagNotFoundError struct {
	TagID int
}

func (e *TagNotFoundError) Error() string {
	return fmt.Sprintf("tag %d not found", e.TagID)
}

// scannable — общий интерфейс для QueryRow и Rows
type scannable interface {
	Scan(dest ...any) error
}

// scanManual забирает одну строку результата в Manual
func scanManual(row scannable) (*models.Manual, error) {
	var manual models.Manual

	err := row.Scan(
		&manual.ID,
		&manual.Title,
		&manual.Author,
		&manual.Content,
		&manual.FilePath,
		&manual.ViewsCount,
		&manual.CreatedAt,
		&manual.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &manual, nil
}
