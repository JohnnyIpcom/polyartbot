package logger

import (
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gopkg.in/yaml.v3"
)

type Config struct {
	zap.Config
}

func (cfg *Config) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return errors.New("unexpected node kind")
	}

	var c zap.Config
	switch node.Tag {
	case "!development":
		c = zap.NewDevelopmentConfig()
	case "!production":
		c = zap.NewProductionConfig()
	case "!!map":
		c = zap.Config{}
	default:
		return fmt.Errorf("unknown tag '%s'", node.Tag)
	}

	if err := node.Decode(&c); err != nil {
		return err
	}

	cfg.Config = c
	return nil
}
