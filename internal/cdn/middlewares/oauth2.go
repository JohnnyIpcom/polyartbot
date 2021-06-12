package middlewares

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/johnnyipcom/polyartbot/internal/cdn/services"
)

func OAuth2(oauth2 services.OAuth2Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !oauth2.Enabled() {
			c.Next()
			return
		}

		token, err := oauth2.ValidationBearerToken(c.Request)
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
