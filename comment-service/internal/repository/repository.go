package repository

import (
	"context"

	"github.com/Thedrogon/blogbish/comment-service/internal/models"
)

type CommentRepository interface {
	// Create creates a new comment
	Create(ctx context.Context, comment *models.Comment) error

	// Get retrieves a comment by ID
	Get(ctx context.Context, id string) (*models.Comment, error)

	// Update updates an existing comment
	Update(ctx context.Context, comment *models.Comment) error

	// Delete soft deletes a comment
	Delete(ctx context.Context, id string) error

	// List retrieves comments based on filters
	List(ctx context.Context, filter *models.CommentFilter) ([]*models.Comment, error)

	// GetChildren retrieves all child comments for a parent comment
	GetChildren(ctx context.Context, parentID string) ([]*models.Comment, error)

	// UpdateStatus updates the comment status (for moderation)
	UpdateStatus(ctx context.Context, id string, status string) error

	// IncrementLikes increments the like count for a comment
	IncrementLikes(ctx context.Context, id string) error

	// IncrementReports increments the report count for a comment
	IncrementReports(ctx context.Context, id string) error

}
