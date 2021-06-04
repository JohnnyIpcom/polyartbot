package config

import "time"

// Reconnect ...
type Reconnect struct {
	Interval   time.Duration `yaml:"interval" default:"500ms"`
	MaxAttempt int           `yaml:"maxAttempt" default:"7200"`
}

type Queue struct {
}

type Exchange struct {
	Type string `yaml:"type" default:"direct"`
}

type Binding struct {
	Exchange   string `yaml:"exchange"`
	Queue      string `yaml:"queue"`
	RoutingKey string `yaml:"routingKey"`
}

type AMQP struct {
	PrefetchCount int                 `yaml:"prefetchCount" default:"1"`
	Exchanges     map[string]Exchange `yaml:"exchanges"`
	Queues        map[string]Queue    `yaml:"queues"`
	Bindings      []Binding           `yaml:"bindings"`
}

// RabbitMQ ...
type RabbitMQ struct {
	URI            string        `yaml:"uri"`
	ConnectionName string        `yaml:"connectionName"`
	NotifyTimeout  time.Duration `yaml:"notifyTimeout" default:"100ms"`
	Reconnect      Reconnect     `yaml:"reconnect"`
	AMQP           AMQP          `yaml:"amqp"`
}
