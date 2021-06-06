package services

import (
	"context"
	"encoding/json"

	"github.com/johnnyipcom/polyartbot/cdn/config"
	"github.com/johnnyipcom/polyartbot/models"
	"github.com/johnnyipcom/polyartbot/rabbitmq"
	"go.uber.org/zap"
)

type RabbitMQService interface {
	Publish(ctx context.Context, image models.RabbitMQImage) error
}

type rabbitMQService struct {
	log  *zap.Logger
	amqp *rabbitmq.AMQP
}

var _ RabbitMQService = &rabbitMQService{}

func NewRabbitMQService(cfg config.Config, amqp *rabbitmq.AMQP, log *zap.Logger) (RabbitMQService, error) {
	return &rabbitMQService{
		log:  log.Named("rabbitMQService"),
		amqp: amqp,
	}, nil
}

func (r *rabbitMQService) Publish(ctx context.Context, image models.RabbitMQImage) error {
	r.log.Info("Publishing file...", zap.Object("image", image))
	body, err := json.Marshal(image)
	if err != nil {
		return err
	}

	p := rabbitmq.Publishing{
		MessageId:   image.FileID,
		ContentType: "application/json",
		Body:        body,
	}

	if image.From != 0 {
		return r.amqp.Publish(ctx, p, "image", "upload")
	}

	if image.To != 0 {
		return r.amqp.Publish(ctx, p, "image", "download")
	}

	return nil
}
