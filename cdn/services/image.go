package services

import (
	"errors"
	"io"
	"mime/multipart"
	"sync"

	"github.com/johnnyipcom/polyartbot/cdn/config"
	"github.com/johnnyipcom/polyartbot/cdn/storage"

	"github.com/johnnyipcom/polyartbot/rabbitmq"

	"go.uber.org/zap"
)

type ImageService interface {
	Upload(multipart.File, multipart.FileHeader) (string, int, error)
	Publish(fileID string) error
	Download(fileID string) ([]byte, error)
	Delete(fileID string) error
}

type imageService struct {
	log       *zap.Logger
	storage   storage.Storage
	rabbitMQ  *rabbitmq.RabbitMQ
	imageAMQP *rabbitmq.AMQP
	once      sync.Once
}

var _ ImageService = &imageService{}

func NewImageService(cfg config.Config, s storage.Storage, r *rabbitmq.RabbitMQ, log *zap.Logger) (ImageService, error) {
	amqpConfig, ok := cfg.RabbitMQ.AMQPs["image.upload"]
	if !ok {
		return nil, errors.New("no valid 'image.upload' config")
	}

	return &imageService{
		log:       log,
		storage:   s,
		rabbitMQ:  r,
		imageAMQP: rabbitmq.NewAMQP(amqpConfig, r, log),
	}, nil
}

func (i *imageService) Upload(file multipart.File, header multipart.FileHeader) (string, int, error) {
	data, err := io.ReadAll(file)
	if err != nil {
		return "", 0, err
	}

	fileID, err := i.storage.Upload(header.Filename, data)
	if err != nil {
		return "", 0, err
	}

	return fileID, len(data), err
}

func (i *imageService) Publish(fileID string) error {
	return i.imageAMQP.Publish(rabbitmq.Message{
		MessageID:   fileID,
		ContentType: "text/plain",
		Body:        []byte(fileID),
	})
}

func (i *imageService) Download(fileID string) ([]byte, error) {
	return i.storage.Download(fileID)
}

func (i *imageService) Delete(fileID string) error {
	return i.storage.Delete(fileID)
}
