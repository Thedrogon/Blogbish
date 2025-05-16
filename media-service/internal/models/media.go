package models

import (
	"time"
)

type Media struct {
	ID          string    `json:"id"`
	UserID      int64     `json:"user_id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	URL         string    `json:"url"`
	Path        string    `json:"path"`
	Metadata    Metadata  `json:"metadata"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Metadata struct {
	Width       int    `json:"width,omitempty"`
	Height      int    `json:"height,omitempty"`
	Format      string `json:"format,omitempty"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	AltText     string `json:"alt_text,omitempty"`
}

type MediaUpload struct {
	UserID      int64    `json:"user_id"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	AltText     string   `json:"alt_text,omitempty"`
	Tags        []string `json:"tags,omitempty"`
}

type MediaResponse struct {
	ID          string    `json:"id"`
	UserID      int64     `json:"user_id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"content_type"`
	Size        int64     `json:"size"`
	URL         string    `json:"url"`
	Metadata    Metadata  `json:"metadata"`
	CreatedAt   time.Time `json:"created_at"`
}

func (m *Media) ToResponse() *MediaResponse {
	return &MediaResponse{
		ID:          m.ID,
		UserID:      m.UserID,
		Filename:    m.Filename,
		ContentType: m.ContentType,
		Size:        m.Size,
		URL:         m.URL,
		Metadata:    m.Metadata,
		CreatedAt:   m.CreatedAt,
	}
}
