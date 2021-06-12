package services

import (
	"context"
	"io"
	"mime/multipart"

	"github.com/johnnyipcom/polyartbot/internal/cdn/config"
	"github.com/johnnyipcom/polyartbot/internal/cdn/storage"

	"github.com/h2non/filetype"
	"go.uber.org/zap"
)

type ImageService interface {
	Upload(ctx context.Context, fileID string, file multipart.File, header multipart.FileHeader, metadata map[string]string) (int, error)
	GetMetadata(ctx context.Context, fileID string) (map[string]string, error)
	Download(ctx context.Context, fileID string) ([]byte, error)
	Delete(ctx context.Context, fileID string) error
}

type imageService struct {
	log     *zap.Logger
	storage storage.Storage
}

var _ ImageService = &imageService{}

func NewImageService(cfg config.Config, s storage.Storage, log *zap.Logger) ImageService {
	return &imageService{
		log:     log.Named("imageService"),
		storage: s,
	}
}

func (i *imageService) Upload(ctx context.Context, fileID string, file multipart.File, header multipart.FileHeader, metadata map[string]string) (int, error) {
	i.log.Info("Uploading files...")
	data, err := io.ReadAll(file)
	if err != nil {
		return 0, err
	}

	kind, err := filetype.Match(data)
	if err != nil {
		return 0, err
	}

	doc := make(map[string]string)
	doc["MIME"] = kind.MIME.Value

	for key, value := range metadata {
		doc[key] = value
	}

	if err := i.storage.Upload(ctx, fileID, header.Filename, data, doc); err != nil {
		return 0, err
	}

	return len(data), err
}

func (i *imageService) GetMetadata(ctx context.Context, fileID string) (map[string]string, error) {
	i.log.Info("Getting metadata from file...", zap.String("fileID", fileID))
	return i.storage.GetMetadata(ctx, fileID)
}

func (i *imageService) Download(ctx context.Context, fileID string) ([]byte, error) {
	i.log.Info("Downloading file...", zap.String("fileID", fileID))
	return i.storage.Download(ctx, fileID)
}

func (i *imageService) Delete(ctx context.Context, fileID string) error {
	i.log.Info("Deleting file...", zap.String("fileID", fileID))
	return i.storage.Delete(ctx, fileID)
}
