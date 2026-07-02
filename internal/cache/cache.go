package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func New(client *redis.Client) *Cache {
	return &Cache{
		client: client,
	}
}

func (c *Cache) Get(context context.Context, key string, dest interface{}) (bool, error) {
	val, err := c.client.Get(context, key).Bytes()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return true, json.Unmarshal(val, dest)
}

func (c *Cache) Set(context context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(context, key, data, ttl).Err()
}

func (c *Cache) Delete(context context.Context, key string) error {
	return c.client.Del(context, key).Err()
}
