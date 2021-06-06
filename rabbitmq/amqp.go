package rabbitmq

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/johnnyipcom/polyartbot/config"
	"github.com/johnnyipcom/polyartbot/utils"

	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type AMQP struct {
	cfg           config.AMQP
	log           *zap.Logger
	connection    *amqp.Connection
	channel       *amqp.Channel
	queues        []string
	notifyClose   chan *amqp.Error
	notifyConfirm chan amqp.Confirmation
	alive         bool
	threads       int
	done          chan struct{}
	isConnected   *utils.AsyncBool
}

var ErrDisconnected = errors.New("disconnected from RabbitMQ, trying to reconnect")

func NewAMQP(cfg config.AMQP, log *zap.Logger) *AMQP {
	threads := runtime.GOMAXPROCS(0)
	if numCPU := runtime.NumCPU(); numCPU > threads {
		threads = numCPU
	}

	return &AMQP{
		cfg:         cfg,
		log:         log.Named("amqp"),
		alive:       true,
		threads:     threads,
		done:        make(chan struct{}),
		queues:      make([]string, 0),
		isConnected: utils.NewAsyncBool(false),
	}
}

func (a *AMQP) Connect(ctx context.Context) error {
	go a.reconnect()
	return nil
}

func (a *AMQP) Disconnect(ctx context.Context) error {
	a.done <- struct{}{}
	return nil
}

func (a *AMQP) reconnect() {
	for a.alive {
		a.isConnected.Notify(false)
		retryCount := 0
		for !a.connect() {
			if !a.alive {
				return
			}

			select {
			case <-a.done:
				return

			case <-time.After(a.cfg.Reconnect):
				a.log.Info("Couldn't connect to rabbitMQ", zap.String("uri", a.cfg.URI), zap.Int("retry", retryCount))
				retryCount++
			}
		}

		a.log.Info("Connected to RabbitMQ server", zap.String("uri", a.cfg.URI))
		select {
		case <-a.done:
			return
		case <-a.notifyClose:
		}
	}
}

func (a *AMQP) connect() bool {
	conn, err := amqp.Dial(a.cfg.URI)
	if err != nil {
		a.log.Info("Connection error", zap.Error(err))
		return false
	}

	channel, err := conn.Channel()
	if err != nil {
		a.log.Info("Connection error", zap.Error(err))
		return false
	}

	if err := channel.Confirm(false); err != nil {
		a.log.Info("Connection error", zap.Error(err))
		return false
	}

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
			a.log.Info("Connection error", zap.Error(err))
			return false
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
			a.log.Info("Connection error", zap.Error(err))
			return false
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
			a.log.Info("Connection error", zap.Error(err))
			return false
		}
	}

	a.connection = conn
	a.channel = channel
	a.notifyClose = make(chan *amqp.Error)
	a.notifyConfirm = make(chan amqp.Confirmation)
	a.channel.NotifyClose(a.notifyClose)
	a.channel.NotifyPublish(a.notifyConfirm)

	a.isConnected.Notify(true)
	return true
}

func (a *AMQP) GetQueues() []string {
	return a.queues
}

func (a *AMQP) Publish(ctx context.Context, p Publishing, exchange string, routingKey string) error {
	if err := <-a.isConnected.Await(ctx, true); err != nil {
		return err
	}

	for {
		if err := a.publish(p, exchange, routingKey); err != nil {
			if errors.Is(err, ErrDisconnected) {
				continue
			}

			return err
		}

		select {
		case c := <-a.notifyConfirm:
			if c.Ack {
				return nil
			}

		case <-ctx.Done():
			return ctx.Err()

		case <-time.After(a.cfg.NotifyTimeout):
		}
	}
}

func (a *AMQP) publish(p Publishing, exchange string, routingKey string) error {
	if !a.isConnected.Get() {
		return ErrDisconnected
	}

	return a.channel.Publish(
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
	)
}

type Consumer interface {
	Consume(d Delivery) error
}

func (a *AMQP) Consume(ctx context.Context, queue string, c Consumer) error {
	if err := <-a.isConnected.Await(ctx, true); err != nil {
		a.log.Error("ERROR 1", zap.Error(err))
		return err
	}

	if err := a.channel.Qos(
		a.cfg.PrefetchCount,
		0,
		false,
	); err != nil {
		a.log.Error("ERROR 2", zap.Error(err))
		return err
	}

	var wg sync.WaitGroup

	var connectionDropped bool
	for i := 0; i < a.threads; i++ {
		msgs, err := a.channel.Consume(
			queue,
			fmt.Sprintf("consumer%d", i),
			false,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			a.log.Error("ERROR 3", zap.Error(err))
			return err
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					a.log.Error("ERROR 4", zap.Error(ctx.Err()))
					return

				case msg, ok := <-msgs:
					if !ok {
						connectionDropped = true
						a.log.Error("ERROR 5")
						return
					}

					body := make([]byte, len(msg.Body))
					copy(body, msg.Body)

					delivery := Delivery{
						MessageId:   msg.MessageId,
						ContentType: msg.ContentType,
						Body:        body,
					}

					if err := c.Consume(delivery); err != nil {
						a.log.Error("Consume error", zap.String("consumer", msg.ConsumerTag), zap.Error(err))
						msg.Reject(false)
					} else {
						msg.Ack(false)
					}
				}
			}
		}()
	}

	wg.Wait()
	if connectionDropped {
		a.log.Error("ERROR 6")
		return ErrDisconnected
	}

	return nil
}
