package cache

import (
	"context"

	"github.com/johnnyipcom/polyartbot/cdn/models"
)

type JWTCache interface {
	SaveTokens(ctx context.Context, username string, access *models.JWTToken, refresh *models.JWTToken) error
	FetchAuth(ctx context.Context, uuid string) (string, error)
	DeleteAuth(ctx context.Context, uuid string) error
}

type Cache interface {
	JWTCache
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
}
