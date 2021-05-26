package config

import (
	"os"

	"github.com/creasty/defaults"
	"gopkg.in/yaml.v3"
)

// Server ...
type Server struct {
	Addr string `yaml:"addr" default:":8080"`
}

// Storage ...
type Storage struct {
	URI     string `yaml:"uri"`
	FilesDB string `yaml:"filesDB"`
}

// Config ...
type Config struct {
	Server  Server  `yaml:"server"`
	Storage Storage `yaml:"storage"`
	Logger  Logger  `yaml:"logger"`
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
