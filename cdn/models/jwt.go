package models

import (
	"time"

	"github.com/google/uuid"
)

type JWTToken struct {
	Token     string
	UUID      string
	ExpiresAt int64
}

func NewToken(expires time.Duration) *JWTToken {
	return &JWTToken{
		UUID:      uuid.New().String(),
		ExpiresAt: time.Now().Add(expires).Unix(),
	}
}
