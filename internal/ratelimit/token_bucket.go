package ratelimit

import (
	"context"
	_ "embed"
	"time"

	"github.com/redis/go-redis/v9"
)

//go:embed lua/token_bucket.lua
var tokenBucketScript string

func NewTokenBucket(client *redis.Client, cfg Config) (*TokenBucket, error) {
	tb := &TokenBucket{
		client: client,
		cfg:    cfg,
	}

	ctx := context.Background()
	sha, err := tb.client.ScriptLoad(ctx, tokenBucketScript).Result()
	if err != nil {
		return nil, err
	}
	tb.sha = sha
	return tb, nil
}

func (t *TokenBucket) Allow(ctx context.Context, key string) (*Result, error) {
	result, err := t.client.EvalSha(ctx, t.sha, []string{key},
		t.cfg.Capacity, t.cfg.RefillRate, int(t.cfg.RefillInterval.Seconds()), time.Now().Unix(),
	).Result()
	if err != nil {
		return nil, err
	}

	vals, ok := result.([]interface{})
	if !ok || len(vals) < 2 {
		return nil, nil
	}

	allowed, _ := vals[0].(int64)
	remaining, _ := vals[1].(int64)

	return &Result{
		Allowed:   allowed == 1,
		Remaining: int(remaining),
	}, nil
}
