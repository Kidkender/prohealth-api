package config

import "os"

type RedisConfig struct {
	Host string
	Port string
	Password string
	DB int
}

func NewRedisConfig() *RedisConfig {
	return &RedisConfig{
		Host:     requireEnv("REDIS_HOST"),
		Port:     requireEnv("REDIS_PORT"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	}
}
