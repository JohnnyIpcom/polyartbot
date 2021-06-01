package config

import (
	"os"
	"time"

	pcfg "github.com/johnnyipcom/polyartbot/config"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

type Token struct {
	Secret  string        `yaml:"secret"`
	Expires time.Duration `yaml:"expires"`
}

type JWT struct {
	Access  Token  `yaml:"access"`
	Refresh Token  `yaml:"refresh"`
	Issuer  string `yaml:"issuer" default:"polyartbot"`
}

// Server ...
type Server struct {
	Addr string `yaml:"addr" default:":8080"`
	JWT  JWT    `yaml:"jwt"`
}

// Mongo ...
type Mongo struct {
	URI    string `yaml:"uri"`
	DBName string `yaml:"dbName" default:"polyartbot"`
}

// Redis ...
type Redis struct {
	URI    string `yaml:"uri"`
	DBName string `yaml:"dbName" default:"polyartbot"`
}

// Config ...
type Config struct {
	Server   Server        `yaml:"server"`
	Mongo    Mongo         `yaml:"mongo"`
	Redis    Redis         `yaml:"redis"`
	RabbitMQ pcfg.RabbitMQ `yaml:"rabbitMQ"`
	Logger   pcfg.Logger   `yaml:"logger"`
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
