package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Thedrogon/blogbish/search-service/internal/models"
	"github.com/Thedrogon/blogbish/search-service/internal/repository"
	"github.com/redis/go-redis/v9"
)

type SearchService struct {
	repo  repository.SearchRepository
	cache *redis.Client
}

func NewSearchService(repo repository.SearchRepository, cache *redis.Client) *SearchService {
	return &SearchService{
		repo:  repo,
		cache: cache,
	}
}

func (s *SearchService) Search(ctx context.Context, req *models.SearchRequest) (*models.SearchResponse, error) {
	// Try to get from cache first
	cacheKey := generateCacheKey(req)
	if cachedResult, err := s.getFromCache(ctx, cacheKey); err == nil {
		if response, ok := cachedResult.(*models.SearchResponse); ok {
			return response, nil
		}
	}

	// If not in cache, perform search
	result, err := s.repo.Search(ctx, req)
	if err != nil {
		return nil, err
	}

	// Cache the result
	go s.cacheResult(context.Background(), cacheKey, result)

	return result, nil
}

func (s *SearchService) Suggest(ctx context.Context, req *models.SuggestionRequest) (*models.SuggestionResponse, error) {
	// Try to get from cache first
	cacheKey := generateSuggestionCacheKey(req)
	if cachedResult, err := s.getFromCache(ctx, cacheKey); err == nil {
		if response, ok := cachedResult.(*models.SuggestionResponse); ok {
			return response, nil
		}
	}

	// If not in cache, get suggestions
	result, err := s.repo.Suggest(ctx, req)
	if err != nil {
		return nil, err
	}

	// Cache the result
	go s.cacheResult(context.Background(), cacheKey, result)

	return result, nil
}

func (s *SearchService) IndexPost(ctx context.Context, post *models.SearchablePost) error {
	return s.repo.IndexPost(ctx, post)
}

func (s *SearchService) IndexComment(ctx context.Context, comment *models.SearchableComment) error {
	return s.repo.IndexComment(ctx, comment)
}

// Helper functions for caching

func (s *SearchService) getFromCache(ctx context.Context, key string) (interface{}, error) {
	data, err := s.cache.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var result interface{}
	if err := json.Unmarshal([]byte(data), &result); err != nil {
		return nil, err
	}

	return result, nil
}

func (s *SearchService) cacheResult(ctx context.Context, key string, result interface{}) error {
	data, err := json.Marshal(result)
	if err != nil {
		return err
	}

	return s.cache.Set(ctx, key, string(data), 15*time.Minute).Err()
}

func generateCacheKey(req *models.SearchRequest) string {
	return fmt.Sprintf("search:%s:%s:%d:%d:%s:%s:%s",
		req.Query,
		req.Type,
		req.From,
		req.Size,
		req.Status,
		strings.Join(req.Tags, ","),
		req.Category,
	)
}

func generateSuggestionCacheKey(req *models.SuggestionRequest) string {
	return fmt.Sprintf("suggest:%s:%s:%d:%s",
		req.Query,
		req.Type,
		req.Limit,
		req.Status,
	)
}
