package rabbitmq

import (
	"time"

	"github.com/johnnyipcom/polyartbot/config"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Connection struct {
	*amqp.Connection

	cfg config.Reconnect
	log *zap.Logger
}

func NewConnection(cfg config.RabbitMQ, c *amqp.Connection, log *zap.Logger) *Connection {
	return &Connection{
		Connection: c,

		cfg: cfg.Reconnect,
		log: log.Named("connection"),
	}
}

func (c *Connection) Channel() (*Channel, error) {
	ch, err := c.Connection.Channel()
	if err != nil {
		return nil, err
	}

	channel := NewChannel(c.cfg, ch, c.log)

	go c.reconnect(channel)
	return channel, nil
}

func (c *Connection) reconnect(channel *Channel) {
	for {
		err, ok := <-channel.NotifyClose(make(chan *amqp.Error))
		if !ok || channel.IsClosed() {
			channel.Close()
			break
		}

		c.log.Debug("channel closed", zap.Error(err))
		for {
			time.Sleep(c.cfg.Interval)

			ch, err := c.Connection.Channel()
			if err == nil {
				c.log.Info("channel reconnected")
				channel.Channel = ch
				break
			}
		}
	}
}
