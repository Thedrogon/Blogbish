package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Thedrogon/blogbish/post-service/internal/models"
	"github.com/redis/go-redis/v9"
)

const (
	postKeyPrefix     = "post:"
	categoryKeyPrefix = "category:"
	defaultExpiration = 24 * time.Hour
)

type Cache interface {
	GetPost(ctx context.Context, slug string) (*models.Post, error)
	SetPost(ctx context.Context, post *models.Post) error
	DeletePost(ctx context.Context, slug string) error
	GetCategory(ctx context.Context, slug string) (*models.Category, error)
	SetCategory(ctx context.Context, category *models.Category) error
	DeleteCategory(ctx context.Context, slug string) error
}

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(addr, password string) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       0,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{client: client}, nil
}

func (c *RedisCache) GetPost(ctx context.Context, slug string) (*models.Post, error) {
	key := postKeyPrefix + slug
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get post from cache: %w", err)
	}

	var post models.Post
	if err := json.Unmarshal(data, &post); err != nil {
		return nil, fmt.Errorf("failed to unmarshal post: %w", err)
	}

	return &post, nil
}

func (c *RedisCache) SetPost(ctx context.Context, post *models.Post) error {
	data, err := json.Marshal(post)
	if err != nil {
		return fmt.Errorf("failed to marshal post: %w", err)
	}

	key := postKeyPrefix + post.Slug
	if err := c.client.Set(ctx, key, data, defaultExpiration).Err(); err != nil {
		return fmt.Errorf("failed to set post in cache: %w", err)
	}

	return nil
}

func (c *RedisCache) DeletePost(ctx context.Context, slug string) error {
	key := postKeyPrefix + slug
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete post from cache: %w", err)
	}
	return nil
}

func (c *RedisCache) GetCategory(ctx context.Context, slug string) (*models.Category, error) {
	key := categoryKeyPrefix + slug
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get category from cache: %w", err)
	}

	var category models.Category
	if err := json.Unmarshal(data, &category); err != nil {
		return nil, fmt.Errorf("failed to unmarshal category: %w", err)
	}

	return &category, nil
}

func (c *RedisCache) SetCategory(ctx context.Context, category *models.Category) error {
	data, err := json.Marshal(category)
	if err != nil {
		return fmt.Errorf("failed to marshal category: %w", err)
	}

	key := categoryKeyPrefix + category.Slug
	if err := c.client.Set(ctx, key, data, defaultExpiration).Err(); err != nil {
		return fmt.Errorf("failed to set category in cache: %w", err)
	}

	return nil
}

func (c *RedisCache) DeleteCategory(ctx context.Context, slug string) error {
	key := categoryKeyPrefix + slug
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete category from cache: %w", err)
	}
	return nil
}
