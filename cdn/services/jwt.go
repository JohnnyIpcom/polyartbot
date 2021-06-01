package services

import (
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/johnnyipcom/polyartbot/cdn/config"
	"github.com/johnnyipcom/polyartbot/cdn/models"
	"go.uber.org/zap"
)

type JWTService interface {
	GenerateTokens(name string, admin bool) (*models.JWTToken, *models.JWTToken, error)
	ExtractTokenMetadata(tokenString string, access bool) (string, string, error)
	VerifyToken(tokenString string, access bool) (*jwt.Token, error)
	TokenValid(tokenString string, access bool) error
}

type jwtAccessClaims struct {
	Name  string `json:"name"`
	Admin bool   `json:"admin"`
	UUID  string `json:"access_uuid"`
	jwt.StandardClaims
}

type jwtRefreshClaims struct {
	Name string `json:"name"`
	UUID string `json:"refresh_uuid"`
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

func (j *jwtService) GenerateTokens(username string, admin bool) (*models.JWTToken, *models.JWTToken, error) {
	j.log.Info("Generating tokens...", zap.String("username", username), zap.Bool("admin", admin))
	var err error

	accessToken := models.NewToken(j.cfg.Access.Expires)
	accessToken.Token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, &jwtAccessClaims{
		Name:  username,
		Admin: admin,
		UUID:  accessToken.UUID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: accessToken.ExpiresAt,
			Issuer:    j.cfg.Issuer,
			IssuedAt:  time.Now().Unix(),
		},
	}).SignedString([]byte(j.cfg.Access.Secret))
	if err != nil {
		return nil, nil, err
	}

	refreshToken := models.NewToken(j.cfg.Refresh.Expires)
	refreshToken.Token, err = jwt.NewWithClaims(jwt.SigningMethodHS256, &jwtRefreshClaims{
		Name: username,
		UUID: refreshToken.UUID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: refreshToken.ExpiresAt,
			Issuer:    j.cfg.Issuer,
			IssuedAt:  time.Now().Unix(),
		},
	}).SignedString([]byte(j.cfg.Refresh.Secret))
	if err != nil {
		return nil, nil, err
	}

	return accessToken, refreshToken, nil
}

func (j *jwtService) ExtractTokenMetadata(tokenString string, access bool) (string, string, error) {
	token, err := j.VerifyToken(tokenString, access)
	if err != nil {
		return "", "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", "", err
	}

	uuidField := "refresh_uuid"
	if access {
		uuidField = "access_uuid"
	}

	uuid, ok := claims[uuidField].(string)
	if !ok {
		return "", "", err
	}

	name, ok := claims["name"].(string)
	if !ok {
		return "", "", err
	}

	return uuid, name, nil
}

func (j *jwtService) VerifyToken(tokenString string, access bool) (*jwt.Token, error) {
	return jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		if access {
			return []byte(j.cfg.Access.Secret), nil
		} else {
			return []byte(j.cfg.Refresh.Secret), nil
		}
	})
}

func (j *jwtService) TokenValid(tokenString string, access bool) error {
	token, err := j.VerifyToken(tokenString, access)
	if err != nil {
		return err
	}

	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		return errors.New("invalid token")
	}

	return nil
}
