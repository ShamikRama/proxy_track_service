package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/shamil/proxy_track_service-1/internal/config"
	"github.com/shamil/proxy_track_service-1/pkg/models"
)

type RedisCache struct {
	client *redis.Client
	config config.RedisConfig
}

func NewRedisCache(cfg config.RedisConfig) (CacheRepository, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &RedisCache{
		client: rdb,
		config: cfg,
	}, nil
}

func (r *RedisCache) Get(ctx context.Context, key string) (interface{}, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}
	return val, nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}

	return r.client.Set(ctx, key, jsonData, ttl).Err()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *RedisCache) GetTrackData(ctx context.Context, trackCode string) (*models.TrackData, error) {
	key := fmt.Sprintf("track:%s", trackCode)

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, err
	}

	var trackData models.TrackData
	if err := json.Unmarshal([]byte(val), &trackData); err != nil {
		return nil, fmt.Errorf("failed to unmarshal track data: %w", err)
	}

	return &trackData, nil
}

func (r *RedisCache) SetTrackData(ctx context.Context, trackCode string, data *models.TrackData, ttl time.Duration) error {
	key := fmt.Sprintf("track:%s", trackCode)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal track data: %w", err)
	}

	return r.client.Set(ctx, key, jsonData, ttl).Err()
}

func (r *RedisCache) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("Redis health check failed: %w", err)
	}

	return nil
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}
