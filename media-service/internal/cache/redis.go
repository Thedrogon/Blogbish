package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Thedrogon/blogbish/media-service/internal/models"
	"github.com/redis/go-redis/v9"
)

const (
	mediaKeyPrefix    = "media:"
	defaultExpiration = 24 * time.Hour
)

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

func (c *RedisCache) GetMedia(ctx context.Context, id string) (*models.Media, error) {
	key := mediaKeyPrefix + id
	data, err := c.client.Get(ctx, key).Bytes()
	if err == redis.Nil {
		return nil, fmt.Errorf("media not found in cache")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get media from cache: %w", err)
	}

	var media models.Media
	if err := json.Unmarshal(data, &media); err != nil {
		return nil, fmt.Errorf("failed to unmarshal media: %w", err)
	}

	return &media, nil
}

func (c *RedisCache) SetMedia(ctx context.Context, media *models.Media) error {
	data, err := json.Marshal(media)
	if err != nil {
		return fmt.Errorf("failed to marshal media: %w", err)
	}

	key := mediaKeyPrefix + media.ID
	if err := c.client.Set(ctx, key, data, defaultExpiration).Err(); err != nil {
		return fmt.Errorf("failed to set media in cache: %w", err)
	}

	return nil
}

func (c *RedisCache) DeleteMedia(ctx context.Context, id string) error {
	key := mediaKeyPrefix + id
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete media from cache: %w", err)
	}
	return nil
}
