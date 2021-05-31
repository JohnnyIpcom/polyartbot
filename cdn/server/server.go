package server

import (
	"context"
	"time"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/johnnyipcom/polyartbot/cdn/config"
	"github.com/johnnyipcom/polyartbot/cdn/controllers"
	"go.uber.org/zap"
)

type Server struct {
	cfg config.Server
	log *zap.Logger

	health controllers.HealthController
	image  controllers.ImageController
	router *gin.Engine
}

func New(cfg config.Config, log *zap.Logger, health controllers.HealthController, image controllers.ImageController) (*Server, error) {
	binding.Validator = new(defaultValidator)

	router := gin.New()
	router.Use(ginzap.Ginzap(log, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(log, true))

	server := &Server{
		cfg:    cfg.Server,
		log:    log,
		health: health,
		image:  image,
		router: router,
	}

	if err := server.initRoutes(); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *Server) initRoutes() error {
	s.router.GET("/health", s.health.Health)

	v1 := s.router.Group("v1")
	v1.POST("/image", s.image.Post)
	v1.GET("/image/:filename", s.image.Get)
	v1.DELETE("/image/:filename", s.image.Delete)

	return nil
}

func (s *Server) Start(ctx context.Context) error {
	go s.router.Run(s.cfg.Addr)
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return nil
}
