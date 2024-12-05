package services

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gosmart/config"
)

var RedisClient *redis.Client

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr: config.GetEnv("REDIS_URL"),
		DB:   0,
	})
}

func LogToRedis(key string, value string) error {
	ctx := context.Background()
	return RedisClient.Set(ctx, key, value, 0).Err()
}
