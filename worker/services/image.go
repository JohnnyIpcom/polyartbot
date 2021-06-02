package services

import (
	"context"

	"github.com/johnnyipcom/polyartbot/client"
	"github.com/johnnyipcom/polyartbot/worker/config"
	"go.uber.org/zap"
)

type ImageService interface {
	Download(fileID string) ([]byte, error)
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, data, err := i.client.GetImage(ctx, fileID)
	if err != nil {
		return nil, err
	}

	i.log.Info("Downloaded file", zap.String("fileID", fileID), zap.Int("bytes", len(data)))
	return data, nil
}
