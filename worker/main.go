package main

import (
	"context"
	"flag"

	"github.com/johnnyipcom/polyartbot/client"
	pcfg "github.com/johnnyipcom/polyartbot/config"
	"github.com/johnnyipcom/polyartbot/logger"
	"github.com/johnnyipcom/polyartbot/rabbitmq"

	"github.com/johnnyipcom/polyartbot/worker/config"
	"github.com/johnnyipcom/polyartbot/worker/consumer"
	"github.com/johnnyipcom/polyartbot/worker/services"

	"go.uber.org/fx"
)

type RegisterParams struct {
	fx.In

	RabbitMQ *rabbitmq.RabbitMQ
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
		fx.Provide(func(cfg config.Config) pcfg.RabbitMQ {
			return cfg.RabbitMQ
		}),
		fx.Provide(func(cfg config.Config) pcfg.Client {
			return cfg.Client
		}),
		fx.Provide(rabbitmq.New),
		fx.Provide(consumer.New),
		fx.Provide(client.New),
		fx.Provide(services.NewImageService),
		fx.Invoke(register),
	).Run()
}
