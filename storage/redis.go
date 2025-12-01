package storage

import (
	"Leaderboard/config"
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var RedisCtx = context.Background()

func InitRedis() error {
	cfg := config.LoadConfig()

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     cfg.GetRedisAddress(),
		Password: "",
		DB:       0,
	})

	_, err := RedisClient.Ping(RedisCtx).Result()
	if err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return nil
}

func CloseRedis() {
	if RedisClient != nil {
		RedisClient.Close()
	}
}
