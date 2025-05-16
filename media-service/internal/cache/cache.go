package cache

import (
	"context"

	"github.com/Thedrogon/blogbish/media-service/internal/models"
)

type Cache interface {
	// GetMedia retrieves media metadata from cache
	GetMedia(ctx context.Context, id string) (*models.Media, error)

	// SetMedia stores media metadata in cache
	SetMedia(ctx context.Context, media *models.Media) error

	// DeleteMedia removes media metadata from cache
	DeleteMedia(ctx context.Context, id string) error
}
