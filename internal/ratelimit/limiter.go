package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Capacity       int
	RefillRate     int
	RefillInterval time.Duration
}

type TokenBucket struct {
	client *redis.Client
	cfg    Config
	sha    string
}

type Result struct {
	Allowed   bool
	Remaining int
}

type Limiter interface {
	Allow(ctx context.Context, key string) (*Result, error)
}
