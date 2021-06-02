package server

import (
	"context"
	"time"

	"github.com/johnnyipcom/polyartbot/cdn/config"
	"github.com/johnnyipcom/polyartbot/cdn/controllers"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"go.uber.org/zap"
)

type Server struct {
	cfg config.Server
	log *zap.Logger

	health controllers.HealthController
	image  controllers.ImageController
	oAuth2 controllers.OAuth2Controller
	router *gin.Engine
}

func New(cfg config.Config, log *zap.Logger, health controllers.HealthController, image controllers.ImageController, o controllers.OAuth2Controller) (*Server, error) {
	binding.Validator = new(defaultValidator)

	router := gin.New()
	router.Use(ginzap.Ginzap(log, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(log, true))

	server := &Server{
		cfg:    cfg.Server,
		log:    log,
		health: health,
		image:  image,
		oAuth2: o,
		router: router,
	}

	if err := server.initRoutes(); err != nil {
		return nil, err
	}

	return server, nil
}

func (s *Server) initRoutes() error {
	s.router.GET("/health", s.health.Health)

	auth := s.router.Group("oauth2")
	{
		auth.POST("/token", s.oAuth2.Token)
		auth.POST("/authorize", s.oAuth2.Authorize)
	}

	cdn := s.router.Group("cdn", s.oAuth2.VerifyMiddleware())
	{
		cdn.POST("/image", s.image.Post)
		cdn.GET("/image/:filename", s.image.Get)
		cdn.DELETE("/image/:filename", s.image.Delete)
	}

	return nil
}

func (s *Server) Start(ctx context.Context) error {
	go s.router.Run(s.cfg.Addr)
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return nil
}
