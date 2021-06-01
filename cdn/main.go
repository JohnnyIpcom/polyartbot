package main

import (
	"context"
	"flag"

	"github.com/johnnyipcom/polyartbot/cdn/config"
	"github.com/johnnyipcom/polyartbot/cdn/controllers"
	"github.com/johnnyipcom/polyartbot/cdn/server"
	"github.com/johnnyipcom/polyartbot/cdn/services"
	"github.com/johnnyipcom/polyartbot/cdn/storage"

	pcfg "github.com/johnnyipcom/polyartbot/config"
	"github.com/johnnyipcom/polyartbot/logger"
	"github.com/johnnyipcom/polyartbot/rabbitmq"

	"go.uber.org/fx"
)

type RegisterParams struct {
	fx.In

	Storage  storage.Storage
	RabbitMQ *rabbitmq.RabbitMQ
	Server   *server.Server
}

func register(lifecycle fx.Lifecycle, p RegisterParams) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := p.Storage.Connect(ctx); err != nil {
				return err
			}

			if err := p.RabbitMQ.Connect(ctx); err != nil {
				return err
			}

			return p.Server.Start(ctx)
		},

		OnStop: func(ctx context.Context) error {
			p.Server.Stop(ctx)
			p.RabbitMQ.Disconnect(ctx)
			return p.Storage.Disconnect(ctx)
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
		//fx.StartTimeout(30*time.Minute), // uncomment this for debug
		fx.Supply(cfg, log),
		fx.Provide(func(cfg config.Config) pcfg.RabbitMQ {
			return cfg.RabbitMQ
		}),
		fx.Provide(storage.NewMongo),
		fx.Provide(rabbitmq.New),
		fx.Provide(controllers.NewHealthController),
		fx.Provide(controllers.NewImageController),
		fx.Provide(services.NewImageService),
		fx.Provide(server.New),
		fx.Invoke(register),
	).Run()
}
