package logger

import (
	"go.uber.org/zap"
)

func New(cfg Config) (*zap.Logger, error) {
	zlog, err := cfg.Config.Build()
	if err != nil {
		return nil, err
	}

	return zlog, nil
}
