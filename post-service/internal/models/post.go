package models

import (
	"time"
)

type Post struct {
	ID          int64     `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	Content     string    `json:"content" db:"content"`
	Slug        string    `json:"slug" db:"slug"`
	AuthorID    int64     `json:"author_id" db:"author_id"`
	CategoryID  int64     `json:"category_id" db:"category_id"`
	Status      string    `json:"status" db:"status"` // draft, published, archived
	Tags        []string  `json:"tags" db:"tags"`
	ViewCount   int64     `json:"view_count" db:"view_count"`
	PublishedAt time.Time `json:"published_at,omitempty" db:"published_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type PostCreate struct {
	Title      string   `json:"title" validate:"required,min=3,max=255"`
	Content    string   `json:"content" validate:"required"`
	CategoryID int64    `json:"category_id" validate:"required"`
	Tags       []string `json:"tags,omitempty"`
	Status     string   `json:"status" validate:"required,oneof=draft published"`
}

type PostUpdate struct {
	Title      string   `json:"title,omitempty" validate:"omitempty,min=3,max=255"`
	Content    string   `json:"content,omitempty"`
	CategoryID int64    `json:"category_id,omitempty"`
	Tags       []string `json:"tags,omitempty"`
	Status     string   `json:"status,omitempty" validate:"omitempty,oneof=draft published archived"`
}

type PostResponse struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Slug        string    `json:"slug"`
	AuthorID    int64     `json:"author_id"`
	CategoryID  int64     `json:"category_id"`
	Status      string    `json:"status"`
	Tags        []string  `json:"tags"`
	ViewCount   int64     `json:"view_count"`
	PublishedAt time.Time `json:"published_at,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type PostFilter struct {
	AuthorID    int64    `json:"author_id,omitempty"`
	CategoryID  int64    `json:"category_id,omitempty"`
	Status      string   `json:"status,omitempty"`
	Tags        []string `json:"tags,omitempty"`
	SearchQuery string   `json:"search_query,omitempty"`
	Page        int      `json:"page,omitempty"`
	PageSize    int      `json:"page_size,omitempty"`
}

func (p *Post) ToResponse() *PostResponse {
	return &PostResponse{
		ID:          p.ID,
		Title:       p.Title,
		Content:     p.Content,
		Slug:        p.Slug,
		AuthorID:    p.AuthorID,
		CategoryID:  p.CategoryID,
		Status:      p.Status,
		Tags:        p.Tags,
		ViewCount:   p.ViewCount,
		PublishedAt: p.PublishedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}
