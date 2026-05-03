package interfaces

import (
	"context"
	"time"
)

type CacheRepository interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string, destination any) error
	Delete(ctx context.Context, key string) error
}
