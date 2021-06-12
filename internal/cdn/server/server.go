package server

import (
	"context"
	"time"

	"github.com/johnnyipcom/polyartbot/internal/cdn/config"
	"github.com/johnnyipcom/polyartbot/internal/cdn/controllers"
	"github.com/johnnyipcom/polyartbot/internal/cdn/middlewares"
	"github.com/johnnyipcom/polyartbot/internal/cdn/services"

	ginzap "github.com/gin-contrib/zap"
	"github.com/gin-gonic/gin"
	"go.uber.org/fx"
	"go.uber.org/zap"
)

type ServerParams struct {
	fx.In

	Cfg config.Config
	Log *zap.Logger

	Health     controllers.HealthController
	Image      controllers.ImageController
	OAuth2     controllers.OAuth2Controller
	OAuth2Serv services.OAuth2Service
}

type Server struct {
	router  *gin.Engine
	address string
}

func New(p ServerParams) *Server {
	router := gin.New()
	router.Use(ginzap.Ginzap(p.Log, time.RFC3339, true))
	router.Use(ginzap.RecoveryWithZap(p.Log, true))
	router.Use(middlewares.Timeout(p.Cfg.Server.Timeout))

	router.GET("/health", p.Health.Health)

	auth := router.Group("oauth2")
	{
		auth.POST("/token", p.OAuth2.Token)
		auth.POST("/authorize", p.OAuth2.Authorize)
	}

	cdn := router.Group("cdn", middlewares.OAuth2(p.OAuth2Serv))
	{
		cdn.POST("/image", p.Image.Post)
		cdn.GET("/image/:filename", p.Image.Get)
		cdn.DELETE("/image/:filename", p.Image.Delete)
	}

	return &Server{
		router:  router,
		address: p.Cfg.Server.Addr,
	}
}

func (s *Server) Start(ctx context.Context) error {
	go s.router.Run(s.address)
	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	return nil
}
