package repository

import (
	"context"

	"github.com/Thedrogon/blogbish/search-service/internal/models"
)

type SearchRepository interface {
	// Index operations
	IndexPost(ctx context.Context, post *models.SearchablePost) error
	IndexComment(ctx context.Context, comment *models.SearchableComment) error

	// Search operations
	Search(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error)
	Suggest(ctx context.Context, req *models.SuggestionRequest) (*models.SuggestionResponse, error)
}
