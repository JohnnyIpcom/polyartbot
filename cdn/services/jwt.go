package services

import (
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/johnnyipcom/polyartbot/cdn/config"
	"go.uber.org/zap"
)

type JWTService interface {
	GenerateToken(name string, admin bool) (string, error)
	ValidateToken(tokenString string) (*jwt.Token, error)
}

// jwtCustomClaims are custom claims extending default ones.
type jwtCustomClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	jwt.StandardClaims
}

type jwtService struct {
	cfg config.JWT
	log *zap.Logger
}

func NewJWTService(cfg config.Config, log *zap.Logger) JWTService {
	return &jwtService{
		cfg: cfg.Server.JWT,
		log: log.Named("jwtService"),
	}
}

func (j *jwtService) GenerateToken(username string, admin bool) (string, error) {
	j.log.Info("Generating token...", zap.String("username", username), zap.Bool("admin", admin))
	claims := &jwtCustomClaims{
		username,
		admin,
		jwt.StandardClaims{
			ExpiresAt: time.Now().Add(j.cfg.Expires).Unix(),
			Issuer:    j.cfg.Issuer,
			IssuedAt:  time.Now().Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	t, err := token.SignedString([]byte(j.cfg.Secret))
	if err != nil {
		return "", err
	}

	return t, nil
}

func (j *jwtService) ValidateToken(tokenString string) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.cfg.Secret), nil
	})
}
