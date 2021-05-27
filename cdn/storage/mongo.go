package storage

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/johnnyipcom/polyartbot/cdn/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.uber.org/zap"
)

type Mongo struct {
	cfg    config.Mongo
	log    *zap.Logger
	client *mongo.Client
	db     *mongo.Database
}

func NewMongo(cfg config.Config, log *zap.Logger) (Storage, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		return nil, err
	}

	return &Mongo{
		cfg:    cfg.Mongo,
		log:    log,
		client: client,
	}, nil
}

func (s *Mongo) Connect(ctx context.Context) error {
	s.log.Info("Connecting to storage...", zap.String("uri", s.cfg.URI))
	if err := s.client.Connect(ctx); err != nil {
		return err
	}

	if err := s.client.Ping(ctx, readpref.Primary()); err != nil {
		return err
	}

	s.db = s.client.Database(s.cfg.DBName)
	return nil
}

func (s *Mongo) Disconnect(ctx context.Context) error {
	s.log.Info("Disconnecting from storage...")
	return s.client.Disconnect(ctx)
}

func (s *Mongo) Upload(name string, p []byte) (string, error) {
	bucket, err := gridfs.NewBucket(s.db)
	if err != nil {
		return "", err
	}

	fileID := uuid.New()
	uploadStream, err := bucket.OpenUploadStreamWithID(fileID, name)
	if err != nil {
		return "", err
	}
	defer uploadStream.Close()

	if _, err := uploadStream.Write(p); err != nil {
		return "", err
	}

	return fileID.String(), nil
}

func (s *Mongo) Download(fileID string) ([]byte, error) {
	bucket, err := gridfs.NewBucket(s.db)
	if err != nil {
		return nil, err
	}

	downloadStream, err := bucket.OpenDownloadStream(uuid.MustParse(fileID))
	if err != nil {
		return nil, err
	}
	defer downloadStream.Close()

	data, err := io.ReadAll(downloadStream)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (s *Mongo) Delete(fileID string) error {
	bucket, err := gridfs.NewBucket(s.db)
	if err != nil {
		return err
	}

	return bucket.Delete(uuid.MustParse(fileID))
}
