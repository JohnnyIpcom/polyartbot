package rabbitmq

import (
	"errors"
	"sync"
	"time"

	"github.com/johnnyipcom/polyartbot/config"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Message struct {
	MessageID   string
	ContentType string
	Body        []byte
}

type AMQP struct {
	cfg      config.AMQP
	log      *zap.Logger
	rabbitMQ *RabbitMQ
	once     sync.Once
}

func NewAMQP(cfg config.AMQP, rabbitMQ *RabbitMQ, log *zap.Logger) *AMQP {
	return &AMQP{
		cfg:      cfg,
		log:      log,
		rabbitMQ: rabbitMQ,
	}
}

func (a *AMQP) initOnce() {
	a.once.Do(func() {
		if err := a.Setup(); err != nil {
			a.log.Fatal("can't init AMQP", zap.Error(err))
		}
	})
}

func (a *AMQP) Setup() error {
	channel, err := a.rabbitMQ.Channel()
	if err != nil {
		return err
	}
	defer channel.Close()

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

	if err := channel.Qos(
		a.cfg.PrefetchCount,
		0,
		false,
	); err != nil {
		return err
	}

	return nil
}

func (a *AMQP) Publish(m Message) error {
	a.initOnce()

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
			MessageId:    m.MessageID,
			ContentType:  m.ContentType,
			Body:         m.Body,
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

func (a *AMQP) Consume(consumer string) (<-chan Message, error) {
	a.initOnce()

	channel, err := a.rabbitMQ.Channel()
	if err != nil {
		return nil, err
	}

	deliveries, err := channel.Consume(
		a.cfg.QueueName,
		consumer,
		false,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		return nil, err
	}

	msgs := make(chan Message)
	go func() {
		for d := range deliveries {
			m := Message{
				MessageID:   d.MessageId,
				ContentType: d.ContentType,
			}

			m.Body = make([]byte, len(d.Body))
			copy(m.Body, d.Body)

			if err := d.Ack(false); err != nil {
				a.log.Error("Unable to acknowledge the message, dropped", zap.Error(err))
			}
		}
	}()

	return msgs, nil
}
