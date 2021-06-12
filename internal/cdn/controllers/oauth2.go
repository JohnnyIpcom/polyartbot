package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/johnnyipcom/polyartbot/internal/cdn/config"
	"github.com/johnnyipcom/polyartbot/internal/cdn/services"
	"go.uber.org/zap"
)

type OAuth2Controller interface {
	Authorize(c *gin.Context)
	Token(c *gin.Context)
}

type oAuth2Controller struct {
	cfg    config.Server
	oauth2 services.OAuth2Service
}

var _ OAuth2Controller = &oAuth2Controller{}

func NewOAuth2Controller(cfg config.Config, oauth2 services.OAuth2Service, log *zap.Logger) OAuth2Controller {
	return &oAuth2Controller{
		cfg:    cfg.Server,
		oauth2: oauth2,
	}
}

func (o *oAuth2Controller) Authorize(c *gin.Context) {
	err := o.oauth2.HandleAuthorizeRequest(c.Writer, c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Abort()
}

func (o *oAuth2Controller) Token(c *gin.Context) {
	err := o.oauth2.HandleTokenRequest(c.Writer, c.Request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.Abort()
}
