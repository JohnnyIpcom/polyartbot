package config

import "time"

// Reconnect ...
type Reconnect struct {
	Interval   time.Duration `yaml:"interval" default:"500ms"`
	MaxAttempt int           `yaml:"maxAttempt" default:"7200"`
}

type AMQP struct {
	ExchangeName string `yaml:"exchangeName"`
	ExchangeType string `yaml:"exchangeType"`
	RoutingKey   string `yaml:"routingKey"`
	QueueName    string `yaml:"queueName"`
}

// RabbitMQ ...
type RabbitMQ struct {
	URI            string          `yaml:"uri"`
	ConnectionName string          `yaml:"connectionName"`
	NotifyTimeout  time.Duration   `yaml:"notifyTimeout" default:"100ms"`
	Reconnect      Reconnect       `yaml:"reconnect"`
	AMQPs          map[string]AMQP `yaml:"AMQPs"`
}
