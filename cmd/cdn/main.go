package main

import (
	"context"
	"flag"

	"github.com/johnnyipcom/polyartbot/internal/cdn/config"
	"github.com/johnnyipcom/polyartbot/internal/cdn/controllers"
	"github.com/johnnyipcom/polyartbot/internal/cdn/server"
	"github.com/johnnyipcom/polyartbot/internal/cdn/services"
	"github.com/johnnyipcom/polyartbot/internal/cdn/storage"

	"github.com/johnnyipcom/polyartbot/pkg/logger"
	"github.com/johnnyipcom/polyartbot/pkg/rabbitmq"

	"go.uber.org/fx"
)

type RegisterParams struct {
	fx.In

	Storage storage.Storage
	AMQP    *rabbitmq.AMQP
	Server  *server.Server
}

func register(lifecycle fx.Lifecycle, p RegisterParams) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := p.Storage.Connect(ctx); err != nil {
				return err
			}

			if err := p.AMQP.Connect(ctx); err != nil {
				return err
			}

			return p.Server.Start(ctx)
		},

		OnStop: func(ctx context.Context) error {
			p.Server.Stop(ctx)
			p.AMQP.Disconnect(ctx)
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
		fx.Provide(func(cfg config.Config) rabbitmq.Config {
			return cfg.RabbitMQ
		}),
		fx.Provide(storage.NewMongo),
		fx.Provide(rabbitmq.NewAMQP),
		fx.Provide(controllers.NewHealthController),
		fx.Provide(controllers.NewImageController),
		fx.Provide(controllers.NewOAuth2Controller),
		fx.Provide(services.NewImageService),
		fx.Provide(services.NewRabbitMQService),
		fx.Provide(services.NewOAuth2Service),
		fx.Provide(server.New),
		fx.Invoke(register),
	).Run()
}
