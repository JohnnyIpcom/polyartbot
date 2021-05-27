package rabbitmq

import (
	"errors"
	"log"
	"time"

	"github.com/johnnyipcom/polyartbot/cdn/config"
	"github.com/streadway/amqp"
)

type AMQP struct {
	cfg      config.AMQP
	rabbitMQ *RabbitMQ
}

func NewAMQP(cfg config.AMQP, rabbitMQ *RabbitMQ) *AMQP {
	return &AMQP{
		cfg:      cfg,
		rabbitMQ: rabbitMQ,
	}
}

func (a AMQP) Setup() error {
	channel, err := a.rabbitMQ.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	if err := a.declareCreate(channel); err != nil {
		return err
	}

	return nil
}

func (a AMQP) Publish(messageID string, contentType string, body []byte) error {
	channel, err := a.rabbitMQ.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	if err := channel.Confirm(false); err != nil {
		return err
	}

	if err := channel.Publish(
		a.cfg.ExchangeName,
		a.cfg.RoutingKey,
		true,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			MessageId:    messageID,
			ContentType:  contentType,
			Body:         body,
		},
	); err != nil {
		return err
	}

	select {
	case ntf := <-channel.NotifyPublish(make(chan amqp.Confirmation, 1)):
		if !ntf.Ack {
			return errors.New("failed to deliver message to exchange/queue")
		}
	case <-channel.NotifyReturn(make(chan amqp.Return)):
		return errors.New("failed to deliver message to exchange/queue")
	case <-time.After(a.rabbitMQ.cfg.NotifyTimeout):
		log.Println("message delivery confirmation to exchange/queue timed out")
	}

	return nil
}

func (a AMQP) declareCreate(channel *amqp.Channel) error {
	if err := channel.ExchangeDeclare(
		a.cfg.ExchangeName,
		a.cfg.ExchangeType,
		true,
		false,
		false,
		false,
		nil,
	); err != nil {
		return err
	}

	if _, err := channel.QueueDeclare(
		a.cfg.QueueName,
		true,
		false,
		false,
		false,
		amqp.Table{"x-queue-mode": "lazy"},
	); err != nil {
		return err
	}

	if err := channel.QueueBind(
		a.cfg.QueueName,
		a.cfg.RoutingKey,
		a.cfg.ExchangeName,
		false,
		nil,
	); err != nil {
		return err
	}

	return nil
}
