package services

import (
	"encoding/json"

	"github.com/johnnyipcom/polyartbot/cdn/config"
	"github.com/johnnyipcom/polyartbot/models"
	"github.com/johnnyipcom/polyartbot/rabbitmq"
	"go.uber.org/zap"
)

type RabbitMQService interface {
	Publish(image models.RabbitMQImage) error
}

type rabbitMQService struct {
	log      *zap.Logger
	rabbitMQ *rabbitmq.RabbitMQ
	amqp     *rabbitmq.AMQP
}

var _ RabbitMQService = &rabbitMQService{}

func NewRabbitMQService(cfg config.Config, r *rabbitmq.RabbitMQ, log *zap.Logger) (RabbitMQService, error) {
	return &rabbitMQService{
		log:      log.Named("rabbitMQService"),
		rabbitMQ: r,
		amqp:     rabbitmq.NewAMQP(cfg.RabbitMQ.AMQP, r, log),
	}, nil
}

func (r *rabbitMQService) Publish(image models.RabbitMQImage) error {
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
		return r.amqp.Publish(p, "image", "upload")
	}

	if image.To != 0 {
		return r.amqp.Publish(p, "image", "download")
	}

	return nil
}
