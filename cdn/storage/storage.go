package storage

import (
	"context"

	"github.com/johnnyipcom/polyartbot/cdn/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

type Storage struct {
	cfg    config.Storage
	log    *zap.Logger
	client *mongo.Client
	dbase  *mongo.Database
	files  *mongo.Collection
}

func New(cfg config.Config, log *zap.Logger) (*Storage, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(cfg.Storage.URI))
	if err != nil {
		return nil, err
	}

	return &Storage{
		cfg:    cfg.Storage,
		log:    log,
		client: client,
	}, nil
}

func (s *Storage) Connect(ctx context.Context) error {
	s.log.Info("Connecting to storage...", zap.String("uri", s.cfg.URI))
	if err := s.client.Connect(ctx); err != nil {
		return err
	}

	if err := s.client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}

	s.dbase = s.client.Database(s.cfg.DBName)
	s.files = s.dbase.Collection("files")
	return nil
}

func (s *Storage) Disconnect(ctx context.Context) error {
	s.log.Info("Disconnecting from storage...")
	return s.client.Disconnect(ctx)
}
