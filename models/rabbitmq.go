package models

import "go.uber.org/zap/zapcore"

type RabbitMQImage struct {
	FileID string `json:"fileID"`
	From   int64  `json:"from"`
}

func NewRabbitMQImage(fileID string, from int64) RabbitMQImage {
	return RabbitMQImage{
		FileID: fileID,
		From:   from,
	}
}

func (r RabbitMQImage) MarshalLogObject(enc zapcore.ObjectEncoder) error {
	enc.AddString("fileID", r.FileID)
	enc.AddInt64("from", r.From)
	return nil
}
