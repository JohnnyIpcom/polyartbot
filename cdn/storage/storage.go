package storage

import (
	"context"
)

type ImageStorage interface {
	Upload(name string, bytes []byte) (string, error)
	Download(fileID string) ([]byte, error)
	Delete(fileID string) error
}

type Storage interface {
	ImageStorage
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
}
