package services

import (
	"context"

	"github.com/johnnyipcom/polyartbot/cdn/cache"
	"github.com/johnnyipcom/polyartbot/cdn/config"
	"github.com/johnnyipcom/polyartbot/cdn/models"
	"go.uber.org/zap"
)

type CacheService interface {
	SaveAuth(username string, access *models.JWTToken, refresh *models.JWTToken) error
	FetchAuth(uuid string) (string, error)
	DeleteAuth(uuid string) error
}

type cacheService struct {
	log   *zap.Logger
	cache cache.Cache
}

var _ CacheService = &cacheService{}

func NewCacheService(cfg config.Config, c cache.Cache, log *zap.Logger) CacheService {
	return &cacheService{
		log:   log.Named("cacheService"),
		cache: c,
	}
}

func (c *cacheService) SaveAuth(username string, access *models.JWTToken, refresh *models.JWTToken) error {
	c.log.Info("Saving auth tokens...", zap.String("username", username))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return c.cache.SaveTokens(ctx, username, access, refresh)
}

func (c *cacheService) FetchAuth(uuid string) (string, error) {
	c.log.Info("Fetching tokens...", zap.String("uuid", uuid))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return c.cache.FetchAuth(ctx, uuid)
}

func (c *cacheService) DeleteAuth(uuid string) error {
	c.log.Info("Fetching tokens...", zap.String("uuid", uuid))
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return c.cache.DeleteAuth(ctx, uuid)
}
