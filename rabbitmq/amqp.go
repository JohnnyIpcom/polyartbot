package rabbitmq

import (
	"errors"
	"time"

	"github.com/johnnyipcom/polyartbot/config"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type AMQP struct {
	cfg      config.AMQP
	log      *zap.Logger
	rabbitMQ *RabbitMQ
	queues   []string
}

func NewAMQP(cfg config.AMQP, rabbitMQ *RabbitMQ, log *zap.Logger) *AMQP {
	return &AMQP{
		cfg:      cfg,
		log:      log.Named("amqp"),
		queues:   make([]string, 0),
		rabbitMQ: rabbitMQ,
	}
}

func (a *AMQP) Setup() error {
	channel, err := a.rabbitMQ.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	for name, exchange := range a.cfg.Exchanges {
		if err := channel.ExchangeDeclare(
			name,
			exchange.Type,
			true,
			false,
			false,
			false,
			nil,
		); err != nil {
			return err
		}
	}

	for name := range a.cfg.Queues {
		q, err := channel.QueueDeclare(
			name,
			true,
			false,
			false,
			false,
			amqp.Table{"x-queue-mode": "lazy"},
		)
		if err != nil {
			return err
		}

		a.queues = append(a.queues, q.Name)
	}

	for _, binding := range a.cfg.Bindings {
		if err := channel.QueueBind(
			binding.Queue,
			binding.RoutingKey,
			binding.Exchange,
			false,
			nil,
		); err != nil {
			return err
		}
	}

	return nil
}

func (a *AMQP) GetQueues() []string {
	return a.queues
}

func (a *AMQP) Publish(p Publishing, exchange string, routingKey string) error {
	channel, err := a.rabbitMQ.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

	if err := channel.Confirm(false); err != nil {
		return err
	}

	if err := channel.Publish(
		exchange,
		routingKey,
		true,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			MessageId:    p.MessageId,
			ContentType:  p.ContentType,
			Body:         p.Body,
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
		a.log.Info("message delivery confirmation to exchange/queue timed out")
	}

	return nil
}

func (a *AMQP) Consume(queue string, consumer string) (<-chan Delivery, error) {
	channel, err := a.rabbitMQ.Channel()
	if err != nil {
		return nil, err
	}

	if err := channel.Qos(
		a.cfg.PrefetchCount,
		0,
		false,
	); err != nil {
		return nil, err
	}

	return channel.Consume(
		queue,
		consumer,
		false,
		false,
		false,
		false,
		nil,
	)
}
