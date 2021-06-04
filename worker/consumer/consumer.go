package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/johnnyipcom/polyartbot/models"
	"github.com/johnnyipcom/polyartbot/rabbitmq"

	"github.com/johnnyipcom/polyartbot/worker/config"
	"github.com/johnnyipcom/polyartbot/worker/services"

	"go.uber.org/zap"
	"gopkg.in/tomb.v2"
)

type Consumer struct {
	cfg       config.Consumer
	log       *zap.Logger
	rabbitMQ  *rabbitmq.RabbitMQ
	amqp      *rabbitmq.AMQP
	imageServ services.ImageService
	polyServ  services.PolyartService
	tomb      tomb.Tomb
}

func New(cfg config.Config, r *rabbitmq.RabbitMQ, i services.ImageService, p services.PolyartService, log *zap.Logger) (*Consumer, error) {
	return &Consumer{
		cfg:       cfg.Consumer,
		log:       log.Named("consumer"),
		rabbitMQ:  r,
		imageServ: i,
		polyServ:  p,
		amqp:      rabbitmq.NewAMQP(cfg.RabbitMQ.AMQP, r, log),
	}, nil
}

func (c *Consumer) Start(ctx context.Context) error {
	c.log.Info("Starting consumer...")
	for i := 0; i < c.cfg.Processors; i++ {
		err := c.processor(fmt.Sprintf("%s%d", c.cfg.ProcessorTag, i))
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Consumer) Stop(ctx context.Context) error {
	c.tomb.Kill(nil)
	return c.tomb.Wait()
}

func (c *Consumer) processor(name string) error {
	c.log.Info("Starting processor...", zap.String("name", name))
	msgs, err := c.amqp.Consume("image_upload", name)
	if err != nil {
		return err
	}

	c.tomb.Go(func() error {
		for {
			select {
			case msg := <-msgs:
				if err := <-c.processMessage(msg); err != nil {
					c.log.Error("can't process message", zap.Error(err))
					msg.Reject(false)
					continue
				}
				msg.Ack(false)

			case <-c.tomb.Dying():
				return c.tomb.Err()
			}
		}
	})

	return nil
}

func (c *Consumer) processMessage(msg rabbitmq.Delivery) <-chan error {
	c.log.Info("Processing message...", zap.String("id", msg.MessageId))
	out := make(chan error)

	go func() {
		defer close(out)

		if msg.ContentType != "application/json" {
			out <- errors.New("unknown content type")
			return
		}

		var image models.RabbitMQImage
		if err := json.Unmarshal(msg.Body, &image); err != nil {
			out <- err
			return
		}

		oldData, err := c.imageServ.Download(image.FileID)
		if err != nil {
			out <- err
			return
		}

		newData, err := c.polyServ.JustCopy(oldData)
		if err != nil {
			out <- err
			return
		}

		uuid, err := c.imageServ.Upload("result.jpg", newData, image.From)
		if err != nil {
			out <- err
			return
		}

		if err := c.imageServ.Delete(image.FileID); err != nil {
			out <- err
			return
		}

		c.log.Info("Got new UUID", zap.String("uuid", uuid))
	}()

	return out
}
