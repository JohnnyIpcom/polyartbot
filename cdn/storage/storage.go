package storage

import (
	"context"
)

type ImageStorage interface {
	Upload(fileID string, name string, bytes []byte, metadata map[string]string) error
	GetMetadata(fileID string) (map[string]string, error)
	Download(fileID string) ([]byte, error)
	Delete(fileID string) error
}

type Storage interface {
	ImageStorage
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
}
