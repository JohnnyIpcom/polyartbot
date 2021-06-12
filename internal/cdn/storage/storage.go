package storage

import (
	"context"
)

type ImageStorage interface {
	Upload(ctx context.Context, fileID string, name string, bytes []byte, metadata map[string]string) error
	GetMetadata(ctx context.Context, fileID string) (map[string]string, error)
	Download(ctx context.Context, fileID string) ([]byte, error)
	Delete(ctx context.Context, fileID string) error
}

type Storage interface {
	ImageStorage
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
}
