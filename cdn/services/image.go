package services

import (
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"

	"github.com/johnnyipcom/polyartbot/cdn/config"
	"github.com/johnnyipcom/polyartbot/cdn/storage"

	"github.com/johnnyipcom/polyartbot/rabbitmq"

	"github.com/h2non/filetype"
	"go.uber.org/zap"
)

type ImageService interface {
	Upload(multipart.File, multipart.FileHeader) (string, int, error)
	Publish(fileID string) error
	GetMetadata(fileID string) (map[string]string, error)
	Download(fileID string) ([]byte, error)
	Delete(fileID string) error
}

type imageService struct {
	log       *zap.Logger
	storage   storage.Storage
	rabbitMQ  *rabbitmq.RabbitMQ
	imageAMQP *rabbitmq.AMQP
}

var _ ImageService = &imageService{}

func NewImageService(cfg config.Config, s storage.Storage, r *rabbitmq.RabbitMQ, log *zap.Logger) (ImageService, error) {
	amqpConfig, ok := cfg.RabbitMQ.AMQPs["image.upload"]
	if !ok {
		return nil, errors.New("no valid 'image.upload' config")
	}

	return &imageService{
		log:       log.Named("imageService"),
		storage:   s,
		rabbitMQ:  r,
		imageAMQP: rabbitmq.NewAMQP(amqpConfig, r, log),
	}, nil
}

func (i *imageService) Upload(file multipart.File, header multipart.FileHeader) (string, int, error) {
	i.log.Info("Uploading files...")
	data, err := io.ReadAll(file)
	if err != nil {
		return "", 0, err
	}

	kind, err := filetype.Match(data)
	if err != nil {
		return "", 0, err
	}

	doc := make(map[string]string)
	doc["MIME"] = kind.MIME.Value

	fileID, err := i.storage.Upload(header.Filename, data, doc)
	if err != nil {
		return "", 0, err
	}

	return fileID, len(data), err
}

func (i *imageService) Publish(fileID string) error {
	i.log.Info("Publishing file...", zap.String("fileID", fileID))

	type uploadImage struct {
		FileID string `json:"fileID"`
	}

	u := uploadImage{FileID: fileID}
	body, err := json.Marshal(u)
	if err != nil {
		return err
	}

	return i.imageAMQP.Publish(rabbitmq.Message{
		MessageID:   fileID,
		ContentType: "application/json",
		Body:        body,
	})
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
