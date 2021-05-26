package storage

import (
	"io"
	"mime/multipart"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo/gridfs"
)

type Image interface {
	Upload(multipart.File, multipart.FileHeader) (string, int, error)
	Download(fileID string) ([]byte, error)
	Delete(fileID string) error
}

func (s *Mongo) Upload(file multipart.File, header multipart.FileHeader) (string, int, error) {
	bucket, err := gridfs.NewBucket(s.files)
	if err != nil {
		return "", 0, err
	}

	fileID := uuid.New()
	uploadStream, err := bucket.OpenUploadStreamWithID(fileID, header.Filename)
	if err != nil {
		return "", 0, err
	}
	defer uploadStream.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return "", 0, err
	}

	size, err := uploadStream.Write(data)
	if err != nil {
		return "", 0, err
	}

	return fileID.String(), size, nil
}

func (s *Mongo) Download(fileID string) ([]byte, error) {
	bucket, err := gridfs.NewBucket(s.files)
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
	bucket, err := gridfs.NewBucket(s.files)
	if err != nil {
		return err
	}

	return bucket.Delete(uuid.MustParse(fileID))
}
