package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"website-of-methodological-materials/internal/models"
)

var ErrTagNameTaken = errors.New("tag name already exists")

type TagRepository struct {
	pool *pgxpool.Pool
}

func NewTagRepository(pool *pgxpool.Pool) *TagRepository {
	return &TagRepository{pool: pool}
}

func (r *TagRepository) Create(ctx context.Context, name string) (*models.Tag, error) {
	const query = `
		INSERT INTO tags (name)
		VALUES ($1)
		RETURNING id, name
	`

	var tag models.Tag
	err := r.pool.QueryRow(ctx, query, name).Scan(&tag.ID, &tag.Name)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrTagNameTaken
		}
		return nil, fmt.Errorf("insert tag: %w", err)
	}

	return &tag, nil
}

func (r *TagRepository) Exists(ctx context.Context, id int) (bool, error) {
	const query = `SELECT EXISTS(SELECT 1 FROM tags WHERE id = $1)`

	var exists bool
	if err := r.pool.QueryRow(ctx, query, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("check tag exists: %w", err)
	}

	return exists, nil
}

func (r *TagRepository) List(ctx context.Context) ([]models.Tag, error) {
	const query = `SELECT id, name FROM tags ORDER BY name`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("select tags: %w", err)
	}
	defer rows.Close()

	tags := make([]models.Tag, 0)
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name); err != nil {
			return nil, fmt.Errorf("scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate tags: %w", err)
	}

	return tags, nil
}
