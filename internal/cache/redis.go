package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewClient(ctx context.Context, redisURL string) (*redis.Client, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}

func Get(ctx context.Context, client *redis.Client, shortcode string) (string, error) {
	return client.Get(ctx, "url:"+shortcode).Result()
}

func Set(ctx context.Context, client *redis.Client, shortcode, longURL string, ttl time.Duration) error {
	return client.Set(ctx, "url:"+shortcode, longURL, ttl).Err()
}

func Delete(ctx context.Context, client *redis.Client, shortcode string) error {
	return client.Del(ctx, "url:"+shortcode).Err()
}
