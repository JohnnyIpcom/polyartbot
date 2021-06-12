package config

import (
	"os"
	"time"

	"github.com/creasty/defaults"
	"github.com/johnnyipcom/polyartbot/pkg/logger"
	"github.com/johnnyipcom/polyartbot/pkg/rabbitmq"
	"gopkg.in/yaml.v3"
)

type OAuth2Client struct {
	ID     string `yaml:"id"`
	Secret string `yaml:"secret"`
	Domain string `yaml:"domain"`
}

type OAuth2 struct {
	Enabled bool           `yaml:"enabled" default:"true"`
	Clients []OAuth2Client `yaml:"clients"`
}

// Server ...
type Server struct {
	Addr    string        `yaml:"addr" default:":8080"`
	Timeout time.Duration `yaml:"timeout" default:"30s"`
	OAuth2  OAuth2        `yaml:"oauth2"`
}

// Mongo ...
type Mongo struct {
	URI    string `yaml:"uri"`
	DBName string `yaml:"dbName"`
}

// Redis ...
type Redis struct {
	URI string `yaml:"uri"`
	DB  int    `yaml:"db"`
}

// Config ...
type Config struct {
	Server   Server          `yaml:"server"`
	Mongo    Mongo           `yaml:"mongo"`
	Redis    Redis           `yaml:"redis"`
	RabbitMQ rabbitmq.Config `yaml:"rabbitmq"`
	Logger   logger.Config   `yaml:"logger"`
}

const (
	DefaultCfgFile = "./cdn.yaml"
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
