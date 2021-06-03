package services

import (
	"github.com/johnnyipcom/polyartbot/client"
	"github.com/johnnyipcom/polyartbot/worker/config"
	"go.uber.org/zap"
)

type ImageService interface {
	Download(fileID string) ([]byte, error)
	Upload(filename string, data []byte, to int64) (string, error)
	Delete(fileID string) error
}

type imageService struct {
	client *client.Client
	log    *zap.Logger
}

var _ ImageService = &imageService{}

func NewImageService(cfg config.Config, c *client.Client, log *zap.Logger) ImageService {
	return &imageService{
		client: c,
		log:    log.Named("imageService"),
	}
}

func (i *imageService) Download(fileID string) ([]byte, error) {
	i.log.Info("Downloading file...", zap.String("fileID", fileID))
	_, data, err := i.client.GetImage(fileID)
	if err != nil {
		return nil, err
	}

	i.log.Info("Downloaded file", zap.String("fileID", fileID), zap.Int("bytes", len(data)))
	return data, nil
}

func (i *imageService) Upload(filename string, data []byte, to int64) (string, error) {
	i.log.Info("Uploading file...", zap.String("filename", filename))
	return i.client.PostImage(filename, data, 0, to)
}

func (i *imageService) Delete(fileID string) error {
	i.log.Info("Deleting file...", zap.String("fileID", fileID))
	return i.client.DeleteImage(fileID)
}
