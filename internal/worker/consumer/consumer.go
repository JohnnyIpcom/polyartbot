package consumer

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/johnnyipcom/polyartbot/pkg/models"
	"github.com/johnnyipcom/polyartbot/pkg/rabbitmq"

	"github.com/johnnyipcom/polyartbot/internal/worker/config"
	"github.com/johnnyipcom/polyartbot/internal/worker/services"

	"go.uber.org/zap"
)

type Consumer struct {
	log       *zap.Logger
	amqp      *rabbitmq.AMQP
	imageServ services.ImageService
	polyServ  services.PolyartService
	cancel    context.CancelFunc
}

func New(cfg config.Config, amqp *rabbitmq.AMQP, i services.ImageService, p services.PolyartService, log *zap.Logger) (*Consumer, error) {
	return &Consumer{
		log:       log.Named("consumer"),
		amqp:      amqp,
		imageServ: i,
		polyServ:  p,
	}, nil
}

func (c *Consumer) Start(context.Context) error {
	ctx, cancel := context.WithCancel(context.Background())
	c.cancel = cancel

	go func() {
		for {
			err := c.amqp.Consume(ctx, "image_upload", c)
			if errors.Is(err, rabbitmq.ErrDisconnected) {
				continue
			}

			c.log.Error("!!!CONSUMER IS DEAD!!!", zap.Error(err))
			break
		}
	}()
	return nil
}

func (c *Consumer) Stop(ctx context.Context) error {
	c.cancel()
	return nil
}

func (c *Consumer) Consume(msg rabbitmq.Delivery) error {
	c.log.Info("Processing message...", zap.String("id", msg.MessageId))
	if msg.ContentType != "application/json" {
		return errors.New("unknown content type")
	}

	var image models.RabbitMQImage
	if err := json.Unmarshal(msg.Body, &image); err != nil {
		return err
	}

	oldData, _, err := c.imageServ.Download(image.FileID)
	if err != nil {
		return err
	}

	newData, err := c.polyServ.JustCopy(oldData)
	if err != nil {
		return err
	}

	uuid, err := c.imageServ.Upload("result.jpg", newData, image.From)
	if err != nil {
		return err
	}

	if err := c.imageServ.Delete(image.FileID); err != nil {
		return err
	}

	c.log.Info("Got new UUID", zap.String("uuid", uuid))
	return nil
}
