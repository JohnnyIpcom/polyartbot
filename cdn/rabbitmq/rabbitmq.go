package rabbitmq

import (
	"errors"
	"sync"
	"time"

	"github.com/johnnyipcom/polyartbot/cdn/config"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type RabbitMQ struct {
	mu         sync.RWMutex
	cfg        config.RabbitMQ
	log        *zap.Logger
	dialConfig amqp.Config
	connection *amqp.Connection
}

func New(cfg config.Config, log *zap.Logger) *RabbitMQ {
	return &RabbitMQ{
		cfg: cfg.RabbitMQ,
		log: log,

		dialConfig: amqp.Config{
			Properties: amqp.Table{
				"connection_name": cfg.RabbitMQ.ConnectionName,
			},
		},
	}
}

func (r *RabbitMQ) Connect() error {
	conn, err := amqp.DialConfig(r.cfg.URI, r.dialConfig)
	if err != nil {
		return err
	}

	r.connection = conn

	go r.reconnect()
	return nil
}

func (r *RabbitMQ) Disconnect() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.connection != nil {
		return r.connection.Close()
	}

	return nil
}

func (r *RabbitMQ) Channel() (*amqp.Channel, error) {
	if r.connection == nil {
		if err := r.Connect(); err != nil {
			return nil, errors.New("connection is not open")
		}
	}

	channel, err := r.connection.Channel()
	if err != nil {
		return nil, err
	}

	return channel, nil
}

func (r *RabbitMQ) Connection() *amqp.Connection {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.connection
}

func (r *RabbitMQ) reconnect() {
WATCH:
	conErr := <-r.connection.NotifyClose(make(chan *amqp.Error))
	if conErr == nil {
		r.log.Info("RabbitMQ connection dropped normally, will not reconnect")
		return
	}

	r.log.Error("RabbitMQ connection dropped, reconnecting")

	var err error
	for i := 1; i <= r.cfg.Reconnect.MaxAttempt; i++ {
		r.mu.RLock()
		r.connection, err = amqp.DialConfig(r.cfg.URI, r.dialConfig)
		r.mu.RUnlock()

		if err == nil {
			r.log.Info("RabbitMQ reconnected")

			goto WATCH
		}

		time.Sleep(r.cfg.Reconnect.Interval)
	}

	r.log.Error("Failed to reconnect to RabbitMQ", zap.Error(err))
}
