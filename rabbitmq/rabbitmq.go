package rabbitmq

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/johnnyipcom/polyartbot/config"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type RabbitMQ struct {
	mu         sync.RWMutex
	cfg        config.RabbitMQ
	log        *zap.Logger
	dialConfig amqp.Config
	connection *Connection
}

func New(cfg config.RabbitMQ, log *zap.Logger) *RabbitMQ {
	return &RabbitMQ{
		cfg: cfg,
		log: log.Named("rabbitMQ"),

		dialConfig: amqp.Config{
			Properties: amqp.Table{
				"connection_name": cfg.ConnectionName,
			},
		},
	}
}

func (r *RabbitMQ) Connect(ctx context.Context) error {
	c, err := amqp.DialConfig(r.cfg.URI, r.dialConfig)
	if err != nil {
		return nil
	}

	r.connection = NewConnection(r.cfg, c, r.log)

	go r.reconnect()
	return nil
}

func (r *RabbitMQ) Disconnect(ctx context.Context) error {
	c, err := r.Connection()
	if err != nil {
		return err
	}

	return c.Close()
}

func (r *RabbitMQ) Connection() (*Connection, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.connection == nil {
		return nil, errors.New("no active connections")
	}

	return r.connection, nil
}

func (r *RabbitMQ) Channel() (*Channel, error) {
	c, err := r.Connection()
	if err != nil {
		return nil, err
	}

	return c.Channel()
}

func (r *RabbitMQ) reconnect() {
	for {
		chanErr, ok := <-r.connection.Connection.NotifyClose(make(chan *amqp.Error))
		if !ok {
			r.log.Debug("connection closed")
			break
		}
		r.log.Info("connection closed", zap.Error(chanErr))

		for {
			time.Sleep(r.cfg.Reconnect.Interval)

			conn, err := amqp.DialConfig(r.cfg.URI, r.dialConfig)
			if err == nil {
				r.mu.Lock()
				r.connection.Connection = conn
				r.mu.Unlock()
				r.log.Info("connection reconnected")
				break
			}
		}
	}
}
