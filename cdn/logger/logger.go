package logger

import (
	"github.com/johnnyipcom/polyartbot/cdn/config"
	"go.uber.org/zap"
)

func New(cfg config.Config) (*zap.Logger, error) {
	zlog, err := cfg.Logger.Config.Build()
	if err != nil {
		return nil, err
	}

	return zlog, nil
}
