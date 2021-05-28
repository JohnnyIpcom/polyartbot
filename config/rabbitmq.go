package config

import (
	"time"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

// Reconnect ...
type Reconnect struct {
	Interval   time.Duration `yaml:"interval" default:"500ms"`
	MaxAttempt int           `yaml:"maxAttempt" default:"7200"`
}

type AMQP struct {
	ExchangeName  string `yaml:"exchangeName"`
	ExchangeType  string `yaml:"exchangeType"`
	RoutingKey    string `yaml:"routingKey"`
	QueueName     string `yaml:"queueName"`
	PrefetchCount int    `yaml:"prefetchCount" default:"1"`
}

func (a *AMQP) UnmarshalYAML(node *yaml.Node) error {
	defaults.MustSet(a)

	type rawAMQP AMQP
	if err := node.Decode((*rawAMQP)(a)); err != nil {
		return err
	}

	return nil
}

// RabbitMQ ...
type RabbitMQ struct {
	URI            string          `yaml:"uri"`
	ConnectionName string          `yaml:"connectionName"`
	NotifyTimeout  time.Duration   `yaml:"notifyTimeout" default:"100ms"`
	Reconnect      Reconnect       `yaml:"reconnect"`
	AMQPs          map[string]AMQP `yaml:"AMQPs"`
}
