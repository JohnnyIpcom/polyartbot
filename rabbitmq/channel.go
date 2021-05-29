package rabbitmq

import (
	"sync/atomic"
	"time"

	"github.com/johnnyipcom/polyartbot/config"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Channel struct {
	*amqp.Channel
	cfg    config.Reconnect
	log    *zap.Logger
	closed int32
}

func NewChannel(cfg config.Reconnect, c *amqp.Channel, log *zap.Logger) *Channel {
	return &Channel{
		Channel: c,

		cfg: cfg,
		log: log.Named("channel"),
	}
}

func (ch *Channel) IsClosed() bool {
	return atomic.LoadInt32(&ch.closed) == 1
}

func (ch *Channel) Close() error {
	if ch.IsClosed() {
		return amqp.ErrClosed
	}

	atomic.StoreInt32(&ch.closed, 1)
	return ch.Channel.Close()
}

func (ch *Channel) Consume(queue, consumer string, autoAck, exclusive, noLocal, noWait bool, args amqp.Table) (<-chan amqp.Delivery, error) {
	deliveries := make(chan amqp.Delivery)

	go func() {
		for {
			d, err := ch.Channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, args)
			if err != nil {
				ch.log.Debug("consume failed", zap.Error(err))
				time.Sleep(ch.cfg.Interval)
				continue
			}

			for msg := range d {
				deliveries <- msg
			}

			time.Sleep(ch.cfg.Interval)
			if ch.IsClosed() {
				break
			}
		}
	}()

	return deliveries, nil
}
