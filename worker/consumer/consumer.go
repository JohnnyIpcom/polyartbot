package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

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
	imageAMQP *rabbitmq.AMQP
	imageServ services.ImageService
	polyServ  services.PolyartService
	tomb      tomb.Tomb
}

func New(cfg config.Config, r *rabbitmq.RabbitMQ, i services.ImageService, p services.PolyartService, log *zap.Logger) (*Consumer, error) {
	amqpConfig, ok := cfg.RabbitMQ.AMQPs["image.upload"]
	if !ok {
		return nil, errors.New("no valid 'image.upload' config")
	}

	return &Consumer{
		cfg:       cfg.Consumer,
		log:       log.Named("consumer"),
		rabbitMQ:  r,
		imageServ: i,
		imageAMQP: rabbitmq.NewAMQP(amqpConfig, r, log),
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
	msgs, err := c.imageAMQP.Consume(name)
	if err != nil {
		return err
	}

	c.tomb.Go(func() error {
		for {
			select {
			case msg := <-msgs:
				if err := <-c.processMessage(msg); err != nil {
					c.log.Error("can't process message", zap.Error(err))
					continue
				}

			case <-c.tomb.Dying():
				return c.tomb.Err()
			}
		}
	})

	return nil
}

func (c *Consumer) processMessage(msg rabbitmq.Message) <-chan error {
	c.log.Info("Processing message...", zap.String("id", msg.MessageID))
	out := make(chan error)

	go func() {
		defer close(out)

		if msg.ContentType != "application/json" {
			out <- errors.New("unknown content type")
			return
		}

		type uploadImage struct {
			FileID string `json:"fileID"`
		}

		var u uploadImage
		if err := json.Unmarshal(msg.Body, &u); err != nil {
			out <- err
			return
		}

		data, err := c.imageServ.Download(u.FileID)
		if err != nil {
			out <- err
			return
		}

		_, err = c.polyServ.Convert(data)
		if err != nil {
			out <- err
			return
		}
	}()

	return out
}
