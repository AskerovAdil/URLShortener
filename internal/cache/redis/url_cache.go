package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/AskerovAdil/URLShortener/internal/domain"
	goredis "github.com/redis/go-redis/v9"
)

type URLCache struct {
	client     *goredis.Client
	defaultTTL time.Duration
}

func NewURLCache(client *goredis.Client, defaultTTL time.Duration) *URLCache {
	return &URLCache{
		client:     client,
		defaultTTL: defaultTTL,
	}
}

func (c *URLCache) Get(ctx context.Context, alias string) (string, error) {
	val, err := c.client.Get(ctx, cacheKey(alias)).Result()
	if errors.Is(err, goredis.Nil) {
		return "", domain.ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("cache get: %w", err)
	}

	return val, nil
}

func (c *URLCache) Set(ctx context.Context, alias, originalURL string, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = c.defaultTTL
	}

	if err := c.client.Set(ctx, cacheKey(alias), originalURL, ttl).Err(); err != nil {
		return fmt.Errorf("cache set: %w", err)
	}

	return nil
}

func (c *URLCache) Delete(ctx context.Context, alias string) error {
	if err := c.client.Del(ctx, cacheKey(alias)).Err(); err != nil {
		return fmt.Errorf("cache delete: %w", err)
	}

	return nil
}

func cacheKey(alias string) string {
	return "url:" + alias
}
