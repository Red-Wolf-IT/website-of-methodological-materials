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
	const query = `
		SELECT id, title, author, content, file_path, views_count, created_at, updated_at
		FROM manuals
		WHERE id = $1
	`

	row := r.pool.QueryRow(ctx, query, id)

	manual, err := scanManual(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, pgx.ErrNoRows
		}
		return nil, fmt.Errorf("select manual by id: %w", err)
	}

	return manual, nil
}

func (r *ManualRepository) List(ctx context.Context) ([]models.Manual, error) {
	const query = `
		SELECT id, title, author, content, file_path, views_count, created_at, updated_at
		FROM manuals
		ORDER BY created_at DESC
	`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("select manuals: %w", err)
	}
	defer rows.Close()

	manuals := make([]models.Manual, 0)
	for rows.Next() {
		manual, err := scanManual(rows)
		if err != nil {
			return nil, fmt.Errorf("scan manual: %w", err)
		}
		manuals = append(manuals, *manual)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate manuals: %w", err)
	}

	return manuals, nil
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
