package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	defaultCacheTTL = 24 * time.Hour
)

type Cache struct {
	log    *slog.Logger
	client *redis.Client
}

func NewRedis(log *slog.Logger, address string) (*Cache, error) {
	const op = "repository.NewRedis"

	opt, err := redis.ParseURL(address)
	if err != nil {
		log.Error("parse address", "address", address, "error", err)
		return nil, fmt.Errorf("%s: parse redis url: %w", op, errors.Join(ErrInternal, err))
	}

	client := redis.NewClient(opt)

	return &Cache{
		log:    log,
		client: client,
	}, nil
}

func (c *Cache) Set(ctx context.Context, shortCode, originalURL string) error {
	const op = "repository.Cache.Set"

	if err := c.client.Set(ctx, shortCode, originalURL, defaultCacheTTL).Err(); err != nil {
		return fmt.Errorf("%s: set key %s: %w", op, shortCode, errors.Join(ErrInternal, err))
	}
	return nil
}

func (c *Cache) Get(ctx context.Context, shortCode string) (string, error) {
	const op = "repository.Cache.Get"

	val, err := c.client.Get(ctx, shortCode).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", fmt.Errorf("%s: key %s: %w", op, shortCode, ErrNotFound)
		}
		return "", fmt.Errorf("%s: key %s: %w", op, shortCode, errors.Join(ErrInternal, err))
	}

	return val, nil
}
