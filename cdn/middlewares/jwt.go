package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/johnnyipcom/polyartbot/cdn/services"
	"github.com/johnnyipcom/polyartbot/cdn/utils"
)

func AuthorizeJWT(j services.JWTService) gin.HandlerFunc {
	return func(c *gin.Context) {
		const BEARER_SCHEMA = "Bearer "
		authHeader := c.GetHeader("Authorization")
		if !strings.HasPrefix(authHeader, BEARER_SCHEMA) {
			restErr := utils.NewUnauthorizedError("no valid authorization bearer", nil)
			c.AbortWithStatusJSON(restErr.Status(), restErr)
			return
		}

		token := authHeader[len(BEARER_SCHEMA):]
		if err := j.TokenValid(token, true); err != nil {
			restErr := utils.NewUnauthorizedError("unauthorized", err)
			c.AbortWithStatusJSON(restErr.Status(), restErr)
		}
	}
}
