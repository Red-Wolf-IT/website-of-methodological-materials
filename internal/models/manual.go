package models

import (
	"time"

	"github.com/google/uuid"
)

type Manual struct {
	ID         uuid.UUID  `json:"id"`
	Title      string     `json:"title"`
	Author     string     `json:"author"`
	Content    string     `json:"content"`
	FilePath   *string    `json:"file_path,omitempty"`
	ViewsCount int        `json:"views_count"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

type ManualCreate struct {
	Title    string
	Author   string
	Content  string
	FilePath *string
}
