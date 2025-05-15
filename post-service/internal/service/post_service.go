package service

import (
	"context"
	"time"

	"github.com/Thedrogon/blogbish/post-service/internal/cache"
	"github.com/Thedrogon/blogbish/post-service/internal/models"
	"github.com/Thedrogon/blogbish/post-service/internal/repository"
	"github.com/Thedrogon/blogbish/post-service/internal/utils"
)

type PostService struct {
	postRepo     repository.PostRepository
	categoryRepo repository.CategoryRepository
	cache        cache.Cache
}

func NewPostService(postRepo repository.PostRepository, categoryRepo repository.CategoryRepository, cache cache.Cache) *PostService {
	return &PostService{
		postRepo:     postRepo,
		categoryRepo: categoryRepo,
		cache:        cache,
	}
}

func (s *PostService) CreatePost(ctx context.Context, input *models.PostCreate, authorID int64) (*models.PostResponse, error) {
	// Validate category exists
	if _, err := s.categoryRepo.GetByID(ctx, input.CategoryID); err != nil {
		return nil, ErrCategoryNotFound
	}

	// Create slug from title
	slug := utils.GenerateUniqueSlug(input.Title, func(slug string) bool {
		_, err := s.postRepo.GetBySlug(ctx, slug)
		return err == nil
	})

	post := &models.Post{
		Title:      input.Title,
		Content:    input.Content,
		Slug:       slug,
		AuthorID:   authorID,
		CategoryID: input.CategoryID,
		Status:     input.Status,
		Tags:       input.Tags,
	}

	// Set PublishedAt if status is published
	if input.Status == "published" {
		now := time.Now()
		post.PublishedAt = now
	}

	if err := s.postRepo.Create(ctx, post); err != nil {
		return nil, err
	}

	return post.ToResponse(), nil
}

func (s *PostService) GetPost(ctx context.Context, slug string) (*models.PostResponse, error) {
	// Try to get from cache first
	if post, err := s.cache.GetPost(ctx, slug); err == nil && post != nil {
		// Increment view count asynchronously
		go func() {
			ctx := context.Background()
			_ = s.postRepo.IncrementViewCount(ctx, post.ID)
		}()
		return post.ToResponse(), nil
	}

	// If not in cache, get from database
	post, err := s.postRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, ErrNotFound
	}

	// Cache the post
	go func() {
		ctx := context.Background()
		_ = s.cache.SetPost(ctx, post)
	}()

	// Increment view count asynchronously
	go func() {
		ctx := context.Background()
		_ = s.postRepo.IncrementViewCount(ctx, post.ID)
	}()

	return post.ToResponse(), nil
}

func (s *PostService) UpdatePost(ctx context.Context, slug string, input *models.PostUpdate, authorID int64) (*models.PostResponse, error) {
	post, err := s.postRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, ErrNotFound
	}

	// Check if user is the author
	if post.AuthorID != authorID {
		return nil, ErrForbidden
	}

	oldSlug := post.Slug

	// Update fields if provided
	if input.Title != "" {
		post.Title = input.Title
		// Generate new slug only if title changes
		newSlug := utils.GenerateUniqueSlug(input.Title, func(slug string) bool {
			if slug == post.Slug {
				return false
			}
			_, err := s.postRepo.GetBySlug(ctx, slug)
			return err == nil
		})
		post.Slug = newSlug
	}

	if input.Content != "" {
		post.Content = input.Content
	}

	if input.CategoryID != 0 {
		if _, err := s.categoryRepo.GetByID(ctx, input.CategoryID); err != nil {
			return nil, ErrCategoryNotFound
		}
		post.CategoryID = input.CategoryID
	}

	if len(input.Tags) > 0 {
		post.Tags = input.Tags
	}

	if input.Status != "" {
		post.Status = input.Status
		// Set PublishedAt if status changes to published
		if input.Status == "published" && post.PublishedAt.IsZero() {
			post.PublishedAt = time.Now()
		}
	}

	if err := s.postRepo.Update(ctx, post); err != nil {
		return nil, err
	}

	// Update cache
	go func() {
		ctx := context.Background()
		// Delete old cache entry if slug changed
		if oldSlug != post.Slug {
			_ = s.cache.DeletePost(ctx, oldSlug)
		}
		_ = s.cache.SetPost(ctx, post)
	}()

	return post.ToResponse(), nil
}

func (s *PostService) DeletePost(ctx context.Context, slug string, authorID int64) error {
	post, err := s.postRepo.GetBySlug(ctx, slug)
	if err != nil {
		return ErrNotFound
	}

	// Check if user is the author
	if post.AuthorID != authorID {
		return ErrForbidden
	}

	if err := s.postRepo.Delete(ctx, post.ID); err != nil {
		return err
	}

	// Delete from cache
	go func() {
		ctx := context.Background()
		_ = s.cache.DeletePost(ctx, slug)
	}()

	return nil
}

func (s *PostService) ListPosts(ctx context.Context, filter *models.PostFilter) ([]*models.PostResponse, error) {
	posts, err := s.postRepo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Convert to response objects
	responses := make([]*models.PostResponse, len(posts))
	for i, post := range posts {
		responses[i] = post.ToResponse()
	}

	return responses, nil
}

func (s *PostService) GetUserPosts(ctx context.Context, authorID int64) ([]*models.PostResponse, error) {
	filter := &models.PostFilter{
		AuthorID: authorID,
	}
	return s.ListPosts(ctx, filter)
}
