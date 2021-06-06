package config

import (
	"os"

	pcfg "github.com/johnnyipcom/polyartbot/config"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

type Polyart struct {
	Steps int `yaml:"steps" default:"300"`
	Shape int `yaml:"shape" default:"1"`
}

// Consumer ...
type Consumer struct {
	Processors   int    `yaml:"processors" default:"1"`
	ProcessorTag string `yaml:"processorsTag" default:"processor"`
}

// Config ...
type Config struct {
	Polyart  Polyart     `yaml:"polyart"`
	Consumer Consumer    `yaml:"consumer"`
	Client   pcfg.Client `yaml:"client"`
	AMQP     pcfg.AMQP   `yaml:"amqp"`
	Logger   pcfg.Logger `yaml:"logger"`
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
