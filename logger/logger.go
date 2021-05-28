package logger

import (
	"github.com/johnnyipcom/polyartbot/config"
	"go.uber.org/zap"
)

func New(cfg config.Logger) (*zap.Logger, error) {
	zlog, err := cfg.Config.Build()
	if err != nil {
		return nil, err
	}

	return zlog, nil
}
