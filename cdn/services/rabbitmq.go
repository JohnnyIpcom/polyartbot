package services

import (
	"encoding/json"
	"errors"

	"github.com/johnnyipcom/polyartbot/cdn/config"
	"github.com/johnnyipcom/polyartbot/models"
	"github.com/johnnyipcom/polyartbot/rabbitmq"
	"go.uber.org/zap"
)

type RabbitMQService interface {
	Publish(image models.RabbitMQImage) error
}

type rabbitMQService struct {
	log       *zap.Logger
	rabbitMQ  *rabbitmq.RabbitMQ
	imageAMQP *rabbitmq.AMQP
}

var _ RabbitMQService = &rabbitMQService{}

func NewRabbitMQService(cfg config.Config, r *rabbitmq.RabbitMQ, log *zap.Logger) (RabbitMQService, error) {
	amqpConfig, ok := cfg.RabbitMQ.AMQPs["image.upload"]
	if !ok {
		return nil, errors.New("no valid 'image.upload' config")
	}

	return &rabbitMQService{
		log:       log.Named("rabbitMQService"),
		rabbitMQ:  r,
		imageAMQP: rabbitmq.NewAMQP(amqpConfig, r, log),
	}, nil
}

func (r *rabbitMQService) Publish(image models.RabbitMQImage) error {
	r.log.Info("Publishing file...", zap.Object("image", image))
	body, err := json.Marshal(image)
	if err != nil {
		return err
	}

	return r.imageAMQP.Publish(rabbitmq.Message{
		MessageID:   image.FileID,
		ContentType: "application/json",
		Body:        body,
	})
}
