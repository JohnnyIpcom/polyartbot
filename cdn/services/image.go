package services

import (
	"io"
	"mime/multipart"

	"github.com/johnnyipcom/polyartbot/cdn/config"
	"github.com/johnnyipcom/polyartbot/cdn/storage"

	"github.com/h2non/filetype"
	"go.uber.org/zap"
)

type ImageService interface {
	Upload(fileID string, file multipart.File, header multipart.FileHeader, metadata map[string]string) (int, error)
	GetMetadata(fileID string) (map[string]string, error)
	Download(fileID string) ([]byte, error)
	Delete(fileID string) error
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

func (i *imageService) Upload(fileID string, file multipart.File, header multipart.FileHeader, metadata map[string]string) (int, error) {
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

	if err := i.storage.Upload(fileID, header.Filename, data, doc); err != nil {
		return 0, err
	}

	return len(data), err
}

func (i *imageService) GetMetadata(fileID string) (map[string]string, error) {
	i.log.Info("Getting metadata from file...", zap.String("fileID", fileID))
	return i.storage.GetMetadata(fileID)
}

func (i *imageService) Download(fileID string) ([]byte, error) {
	i.log.Info("Downloading file...", zap.String("fileID", fileID))
	return i.storage.Download(fileID)
}

func (i *imageService) Delete(fileID string) error {
	i.log.Info("Deleting file...", zap.String("fileID", fileID))
	return i.storage.Delete(fileID)
}
