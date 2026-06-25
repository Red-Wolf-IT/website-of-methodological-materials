package postgres

import (
	"context"
	"fmt"
	"strings"

	"website-of-methodological-materials/internal/models"
)

type sqlBuilder struct {
	parts []string
	args  []any
}

func newSQLBuilder(base string) *sqlBuilder {
	return &sqlBuilder{parts: []string{base}}
}

func (b *sqlBuilder) addPart(part string) {
	b.parts = append(b.parts, part)
}

func (b *sqlBuilder) addArg(value any) string {
	b.args = append(b.args, value)
	return fmt.Sprintf("$%d", len(b.args))
}

func (b *sqlBuilder) String() string {
	return strings.Join(b.parts, " ")
}

func (r *ManualRepository) List(ctx context.Context, filter models.ManualListFilter) (*models.ManualListResult, error) {
	from := newSQLBuilder("FROM manuals m")
	where := make([]string, 0, 3)

	if filter.TagID != nil {
		from.addPart("INNER JOIN manual_tags mt ON mt.manual_id = m.id")
		ph := from.addArg(*filter.TagID)
		where = append(where, fmt.Sprintf("mt.tag_id = %s", ph))
	}

	if filter.Author != "" {
		ph := from.addArg("%" + filter.Author + "%")
		where = append(where, fmt.Sprintf("m.author ILIKE %s", ph))
	}

	if filter.Q != "" {
		ph := from.addArg("%" + filter.Q + "%")
		where = append(where, fmt.Sprintf("(m.title ILIKE %s OR m.content ILIKE %s OR m.author ILIKE %s)", ph, ph, ph))
	}

	whereClause := ""
	if len(where) > 0 {
		whereClause = "WHERE " + strings.Join(where, " AND ")
	}

	countQuery := fmt.Sprintf("SELECT COUNT(DISTINCT m.id) %s %s", from.String(), whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, from.args...).Scan(&total); err != nil {
		return nil, fmt.Errorf("count manuals: %w", err)
	}

	orderBy := "ORDER BY m.created_at DESC"
	if filter.Sort == "popular" {
		orderBy = "ORDER BY m.views_count DESC, m.created_at DESC"
	}

	offset := (filter.Page - 1) * filter.Limit
	limitPH := from.addArg(filter.Limit)
	offsetPH := from.addArg(offset)

	selectQuery := fmt.Sprintf(`
		SELECT DISTINCT m.id, m.title, m.author, m.content, m.file_path, m.views_count, m.created_at, m.updated_at
		%s %s
		%s
		LIMIT %s OFFSET %s
	`, from.String(), whereClause, orderBy, limitPH, offsetPH)

	rows, err := r.pool.Query(ctx, selectQuery, from.args...)
	if err != nil {
		return nil, fmt.Errorf("select manuals: %w", err)
	}
	defer rows.Close()

	items := make([]models.Manual, 0)
	for rows.Next() {
		manual, err := scanManual(rows)
		if err != nil {
			return nil, fmt.Errorf("scan manual: %w", err)
		}
		items = append(items, *manual)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate manuals: %w", err)
	}

	return &models.ManualListResult{
		Items: items,
		Total: total,
		Page:  filter.Page,
		Limit: filter.Limit,
	}, nil
}
