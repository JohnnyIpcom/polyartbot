package cache

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/johnnyipcom/polyartbot/cdn/config"
	"github.com/johnnyipcom/polyartbot/cdn/models"
	"go.uber.org/zap"
)

type Redis struct {
	cfg   config.Redis
	log   *zap.Logger
	redis *redis.Client
}

func NewRedis(cfg config.Config, log *zap.Logger) (Cache, error) {
	redis := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.URI,
		Password: "",
		DB:       0,
	})

	return &Redis{
		cfg:   cfg.Redis,
		log:   log.Named("redis"),
		redis: redis,
	}, nil
}

func (r *Redis) Connect(ctx context.Context) error {
	r.log.Info("Connecting to redis...", zap.String("uri", r.cfg.URI))
	return r.redis.Ping(ctx).Err()
}

func (r *Redis) Disconnect(ctx context.Context) error {
	r.log.Info("Disconnecting from redis...")
	return nil
}

func (r *Redis) SaveTokens(ctx context.Context, username string, access *models.JWTToken, refresh *models.JWTToken) error {
	r.log.Info("Saving tokens...")
	now := time.Now()
	aAt := time.Unix(access.ExpiresAt, 0).Sub(now)
	rAt := time.Unix(refresh.ExpiresAt, 0).Sub(now)

	if err := r.redis.Set(ctx, access.UUID, username, aAt).Err(); err != nil {
		return err
	}

	if err := r.redis.Set(ctx, refresh.UUID, username, rAt).Err(); err != nil {
		return err
	}

	return nil
}

func (r *Redis) FetchAuth(ctx context.Context, uuid string) (string, error) {
	r.log.Info("Fething auth uuid...")
	return r.redis.Get(ctx, uuid).Result()
}

func (r *Redis) DeleteAuth(ctx context.Context, uuid string) error {
	r.log.Info("Fething auth uuid...")
	_, err := r.redis.Del(ctx, uuid).Result()
	return err
}
