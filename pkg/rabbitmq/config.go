package rabbitmq

import "time"

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

type Config struct {
	URI            string              `yaml:"uri"`
	ConnectionName string              `yaml:"connectionName"`
	NotifyTimeout  time.Duration       `yaml:"notifyTimeout" default:"100ms"`
	Reconnect      time.Duration       `yaml:"reconnect" default:"3000ms"`
	PrefetchCount  int                 `yaml:"prefetchCount" default:"1"`
	Exchanges      map[string]Exchange `yaml:"exchanges"`
	Queues         map[string]Queue    `yaml:"queues"`
	Bindings       []Binding           `yaml:"bindings"`
}
