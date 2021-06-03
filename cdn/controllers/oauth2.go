package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	oredis "github.com/go-oauth2/redis/v4"
	"github.com/go-redis/redis/v8"
	"github.com/johnnyipcom/polyartbot/cdn/config"
	"go.uber.org/zap"
)

type OAuth2Controller interface {
	Authorize(c *gin.Context)
	Token(c *gin.Context)

	VerifyMiddleware() gin.HandlerFunc
}

type oAuth2Controller struct {
	cfg    config.Server
	server *server.Server
}

var _ OAuth2Controller = &oAuth2Controller{}

func NewOAuth2Controller(cfg config.Config, log *zap.Logger) OAuth2Controller {
	manager := manage.NewDefaultManager()
	manager.MapTokenStorage(oredis.NewRedisStore(&redis.Options{
		Addr: cfg.Redis.URI,
		DB:   cfg.Redis.DB,
	}))

	clientStore := store.NewClientStore()
	for _, clientCfg := range cfg.Server.OAuth2.Clients {
		clientStore.Set(clientCfg.ID, &models.Client{
			ID:     clientCfg.ID,
			Secret: clientCfg.Secret,
			Domain: clientCfg.Domain,
		})
	}
	manager.MapClientStorage(clientStore)

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)

	return &oAuth2Controller{
		cfg:    cfg.Server,
		server: server.NewDefaultServer(manager),
	}
}

func (o *oAuth2Controller) Authorize(c *gin.Context) {
	err := o.server.HandleAuthorizeRequest(c.Writer, c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Abort()
}

func (o *oAuth2Controller) Token(c *gin.Context) {
	err := o.server.HandleTokenRequest(c.Writer, c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Abort()
}

func (o *oAuth2Controller) VerifyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !o.cfg.OAuth2.Enabled {
			c.Next()
			return
		}

		token, err := o.server.ValidationBearerToken(c.Request)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": err.Error(),
			})
			return
		}

		c.Set("oauth2.token", token)
		c.Next()
	}
}
