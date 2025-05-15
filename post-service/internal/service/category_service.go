package service

import (
	"context"

	"github.com/Thedrogon/blogbish/post-service/internal/cache"
	"github.com/Thedrogon/blogbish/post-service/internal/models"
	"github.com/Thedrogon/blogbish/post-service/internal/repository"
	"github.com/Thedrogon/blogbish/post-service/internal/utils"
)

type CategoryService struct {
	categoryRepo repository.CategoryRepository
	cache        cache.Cache
}

func NewCategoryService(categoryRepo repository.CategoryRepository, cache cache.Cache) *CategoryService {
	return &CategoryService{
		categoryRepo: categoryRepo,
		cache:        cache,
	}
}

func (s *CategoryService) CreateCategory(ctx context.Context, input *models.CategoryCreate) (*models.CategoryResponse, error) {
	// Generate slug from name
	slug := utils.GenerateUniqueSlug(input.Name, func(slug string) bool {
		_, err := s.categoryRepo.GetBySlug(ctx, slug)
		return err == nil
	})

	category := &models.Category{
		Name:        input.Name,
		Slug:        slug,
		Description: input.Description,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, err
	}

	// Cache the new category
	go func() {
		ctx := context.Background()
		_ = s.cache.SetCategory(ctx, category)
	}()

	return category.ToResponse(), nil
}

func (s *CategoryService) GetCategory(ctx context.Context, slug string) (*models.CategoryResponse, error) {
	// Try to get from cache first
	if category, err := s.cache.GetCategory(ctx, slug); err == nil && category != nil {
		return category.ToResponse(), nil
	}

	// If not in cache, get from database
	category, err := s.categoryRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, ErrNotFound
	}

	// Cache the category
	go func() {
		ctx := context.Background()
		_ = s.cache.SetCategory(ctx, category)
	}()

	return category.ToResponse(), nil
}

func (s *CategoryService) UpdateCategory(ctx context.Context, slug string, input *models.CategoryUpdate) (*models.CategoryResponse, error) {
	category, err := s.categoryRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, ErrNotFound
	}

	oldSlug := category.Slug

	if input.Name != "" {
		category.Name = input.Name
		// Generate new slug only if name changes
		newSlug := utils.GenerateUniqueSlug(input.Name, func(slug string) bool {
			if slug == category.Slug {
				return false
			}
			_, err := s.categoryRepo.GetBySlug(ctx, slug)
			return err == nil
		})
		category.Slug = newSlug
	}

	if input.Description != "" {
		category.Description = input.Description
	}

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, err
	}

	// Update cache
	go func() {
		ctx := context.Background()
		// Delete old cache entry if slug changed
		if oldSlug != category.Slug {
			_ = s.cache.DeleteCategory(ctx, oldSlug)
		}
		_ = s.cache.SetCategory(ctx, category)
	}()

	return category.ToResponse(), nil
}

func (s *CategoryService) DeleteCategory(ctx context.Context, slug string) error {
	category, err := s.categoryRepo.GetBySlug(ctx, slug)
	if err != nil {
		return ErrNotFound
	}

	if err := s.categoryRepo.Delete(ctx, category.ID); err != nil {
		return err
	}

	// Delete from cache
	go func() {
		ctx := context.Background()
		_ = s.cache.DeleteCategory(ctx, slug)
	}()

	return nil
}

func (s *CategoryService) ListCategories(ctx context.Context) ([]*models.CategoryResponse, error) {
	categories, err := s.categoryRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Convert to response objects
	responses := make([]*models.CategoryResponse, len(categories))
	for i, category := range categories {
		responses[i] = category.ToResponse()
	}

	return responses, nil
}
