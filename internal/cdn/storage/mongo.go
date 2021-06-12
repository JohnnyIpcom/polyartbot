package storage

import (
	"context"
	"io"

	"github.com/johnnyipcom/polyartbot/internal/cdn/config"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/x/bsonx"
	"go.mongodb.org/mongo-driver/x/bsonx/bsoncore"
	"go.uber.org/zap"
)

type Mongo struct {
	cfg    config.Mongo
	log    *zap.Logger
	client *mongo.Client
	db     *mongo.Database
	bucket *gridfs.Bucket
}

func NewMongo(cfg config.Config, log *zap.Logger) (Storage, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(cfg.Mongo.URI))
	if err != nil {
		return nil, err
	}

	return &Mongo{
		cfg:    cfg.Mongo,
		log:    log.Named("mongo"),
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

	bucket, err := gridfs.NewBucket(s.db)
	if err != nil {
		return err
	}

	s.bucket = bucket
	return nil
}

func (s *Mongo) Disconnect(ctx context.Context) error {
	s.log.Info("Disconnecting from storage...")
	return s.client.Disconnect(ctx)
}

func (s *Mongo) Upload(ctx context.Context, fileID string, name string, p []byte, doc map[string]string) error {
	metadata := make(bsonx.Doc, 0)
	for key, val := range doc {
		metadata = metadata.Append(key, bsonx.String(val))
	}

	opts := options.GridFSUpload()
	opts.SetMetadata(metadata)

	uploadStream, err := s.bucket.OpenUploadStreamWithID(uuid.MustParse(fileID), name, opts)
	if err != nil {
		return err
	}
	defer uploadStream.Close()

	if _, err := uploadStream.Write(p); err != nil {
		return err
	}

	return nil
}

func (s *Mongo) GetMetadata(ctx context.Context, fileID string) (map[string]string, error) {
	id, err := convertFileID(uuid.MustParse(fileID))
	if err != nil {
		return nil, err
	}

	res := s.bucket.GetFilesCollection().FindOne(ctx, bsonx.Doc{{Key: "_id", Value: id}})
	if res.Err() != nil {
		return nil, res.Err()
	}

	type metadataOwner struct {
		Metadata map[string]string `json:"metadata"`
	}

	var m metadataOwner
	if err := res.Decode(&m); err != nil {
		return nil, err
	}

	return m.Metadata, nil
}

func (s *Mongo) Download(ctx context.Context, fileID string) ([]byte, error) {
	s.log.Info("Downloading file...", zap.String("fileID", fileID))
	downloadStream, err := s.bucket.OpenDownloadStream(uuid.MustParse(fileID))
	if err != nil {
		return nil, err
	}
	defer downloadStream.Close()

	data, err := io.ReadAll(downloadStream)
	if err != nil {
		return nil, err
	}

	s.log.Info("Downloaded bytes...", zap.Int("len", len(data)))
	return data, nil
}

func (s *Mongo) Delete(ctx context.Context, fileID string) error {
	s.log.Info("Deleting file...", zap.String("fileID", fileID))
	return s.bucket.Delete(uuid.MustParse(fileID))
}

type _convertFileID struct {
	ID interface{} `bson:"_id"`
}

func convertFileID(fileID interface{}) (bsonx.Val, error) {
	id := _convertFileID{
		ID: fileID,
	}

	b, err := bson.Marshal(id)
	if err != nil {
		return bsonx.Val{}, err
	}

	val := bsoncore.Document(b).Lookup("_id")

	var res bsonx.Val
	err = res.UnmarshalBSONValue(val.Type, val.Data)

	return res, err
}
