package main

import (
	"context"
	"flag"

	"github.com/johnnyipcom/polyartbot/cdn/config"
	"github.com/johnnyipcom/polyartbot/cdn/logger"
	"github.com/johnnyipcom/polyartbot/cdn/server"
	"github.com/johnnyipcom/polyartbot/cdn/storage"
	"go.uber.org/fx"
)

type RegisterParams struct {
	fx.In

	Storage *storage.Storage
	Server  *server.Server
}

func register(lifecycle fx.Lifecycle, p RegisterParams) {
	lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if err := p.Storage.Connect(ctx); err != nil {
				return err
			}

			return p.Server.Start(ctx)
		},

		OnStop: func(ctx context.Context) error {
			if err := p.Server.Stop(ctx); err != nil {
				return err
			}

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

	fx.New(
		//fx.StartTimeout(30*time.Minute), // uncomment this for debug
		fx.Supply(cfg),
		fx.Provide(logger.New),
		fx.Provide(storage.New),
		fx.Provide(server.New),
		fx.Invoke(register),
	).Run()
}
