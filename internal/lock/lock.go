package lock

import (
	"context"
	_ "embed"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

//go:embed lua/unlock.lua
var unlockScript string

type DistributedLock struct {
	client    *redis.Client
	unlockSHA string
}

func New(client *redis.Client) (*DistributedLock, error) {
	sha, err := client.ScriptLoad(context.Background(), unlockScript).Result()
	if err != nil {
		return nil, err
	}
	return &DistributedLock{
		client:    client,
		unlockSHA: sha,
	}, nil
}

func (l *DistributedLock) Acquire(ctx context.Context, key string, ttl time.Duration) (string, error) {
	lockValue := uuid.NewString()
	ok, err := l.client.SetNX(ctx, key, lockValue, ttl).Result()
	if err != nil {
		return "", err
	}
	if !ok {
		return "", nil
	}
	return lockValue, nil
}

func (l *DistributedLock) Release(ctx context.Context, key, lockValue string) error {
	_, err := l.client.EvalSha(ctx, l.unlockSHA, []string{key}, lockValue).Result()
	return err
}
