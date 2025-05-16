package models

import "time"

// SearchablePost represents a post document in Elasticsearch
type SearchablePost struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Content     string    `json:"content"`
	Excerpt     string    `json:"excerpt"`
	AuthorID    int64     `json:"author_id"`
	AuthorName  string    `json:"author_name"`
	Categories  []string  `json:"categories"`
	Tags        []string  `json:"tags"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	PublishedAt time.Time `json:"published_at,omitempty"`
}

// SearchableComment represents a comment document in Elasticsearch
type SearchableComment struct {
	ID        string    `json:"id"`
	PostID    string    `json:"post_id"`
	Content   string    `json:"content"`
	UserID    int64     `json:"user_id"`
	UserName  string    `json:"user_name"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SearchRequest represents a search query
type SearchRequest struct {
	Query     string   `json:"query" binding:"required"`
	Type      string   `json:"type" binding:"required,oneof=post comment all"` // post, comment, or all
	From      int      `json:"from" binding:"min=0"`
	Size      int      `json:"size" binding:"min=1,max=100"`
	Status    string   `json:"status,omitempty"`
	Tags      []string `json:"tags,omitempty"`
	Category  string   `json:"category,omitempty"`
	SortBy    string   `json:"sort_by,omitempty" binding:"omitempty,oneof=relevance date"`
	SortOrder string   `json:"sort_order,omitempty" binding:"omitempty,oneof=asc desc"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Total    int64       `json:"total"`
	From     int         `json:"from"`
	Size     int         `json:"size"`
	Posts    interface{} `json:"posts,omitempty"`
	Comments interface{} `json:"comments,omitempty"`
}

// SuggestionRequest represents a suggestion query
type SuggestionRequest struct {
	Query  string `json:"query" binding:"required"`
	Type   string `json:"type" binding:"required,oneof=post comment tag category"` // post, comment, tag, or category
	Limit  int    `json:"limit" binding:"min=1,max=10"`
	Status string `json:"status,omitempty"`
}

// SuggestionResponse represents a suggestion response
type SuggestionResponse struct {
	Suggestions []string `json:"suggestions"`
}
