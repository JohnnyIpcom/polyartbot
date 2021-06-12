package config

import (
	"os"

	"github.com/creasty/defaults"
	"github.com/johnnyipcom/polyartbot/pkg/client"
	"github.com/johnnyipcom/polyartbot/pkg/logger"
	"github.com/johnnyipcom/polyartbot/pkg/rabbitmq"
	"gopkg.in/yaml.v3"
)

type Polyart struct {
	Steps int `yaml:"steps" default:"300"`
	Shape int `yaml:"shape" default:"1"`
}

// Config ...
type Config struct {
	Polyart  Polyart         `yaml:"polyart"`
	Client   client.Config   `yaml:"client"`
	RabbitMQ rabbitmq.Config `yaml:"rabbitmq"`
	Logger   logger.Config   `yaml:"logger"`
}

const (
	DefaultCfgFile = "./worker.yaml"
)

// Option ...
type Option func(c *Config) error

// NewFromFile ...
func NewFromFile(file string, opts ...Option) (Config, error) {
	var config = Config{}
	if err := config.init(file, opts...); err != nil {
		return config, err
	}

	return config, nil
}

func (config *Config) init(file string, opts ...Option) error {
	bytes, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	defaults.MustSet(config)

	err = yaml.Unmarshal(bytes, config)
	if err != nil {
		return err
	}

	for _, option := range opts {
		if err := option(config); err != nil {
			return err
		}
	}

	return nil
}
