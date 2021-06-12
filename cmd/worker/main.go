package main

import (
	"context"
	"flag"

	"github.com/johnnyipcom/polyartbot/pkg/client"
	"github.com/johnnyipcom/polyartbot/pkg/logger"
	"github.com/johnnyipcom/polyartbot/pkg/rabbitmq"

	"github.com/johnnyipcom/polyartbot/internal/worker/config"
	"github.com/johnnyipcom/polyartbot/internal/worker/consumer"
	"github.com/johnnyipcom/polyartbot/internal/worker/services"

	"go.uber.org/fx"
)

type RegisterParams struct {
	fx.In

	RabbitMQ *rabbitmq.AMQP
	Consumer *consumer.Consumer
}

func register(lifecycle fx.Lifecycle, p RegisterParams) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := p.RabbitMQ.Connect(ctx); err != nil {
				return err
			}

			return p.Consumer.Start(ctx)
		},

		OnStop: func(ctx context.Context) error {
			p.RabbitMQ.Disconnect(ctx)
			return p.Consumer.Stop(ctx)
		},
	})
}

func main() {
	var configFile string
	flag.StringVar(&configFile, "config", config.DefaultCfgFile, "path to config file")
	flag.Parse()

	cfg, err := config.NewFromFile(configFile)
	if err != nil {
		panic(err)
	}

	log, err := logger.New(cfg.Logger)
	if err != nil {
		panic(err)
	}

	fx.New(
		fx.Supply(cfg, log),
		fx.Provide(func(cfg config.Config) rabbitmq.Config {
			return cfg.RabbitMQ
		}),
		fx.Provide(func(cfg config.Config) client.Config {
			return cfg.Client
		}),
		fx.Provide(rabbitmq.NewAMQP),
		fx.Provide(consumer.New),
		fx.Provide(client.New),
		fx.Provide(services.NewImageService),
		fx.Provide(services.NewPolyartService),
		fx.Invoke(register),
	).Run()
}
