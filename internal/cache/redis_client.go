package cache

import (
	"context"
	"log/slog"
	"time"

	"backend-path/internal/config"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, cfg config.Config, logg *slog.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx).Err(); err != nil {
		logg.Error("redis connection failed", "err", err)
		return client
	}

	logg.Info("redis connected", "addr", cfg.RedisAddr)

	return client
}
