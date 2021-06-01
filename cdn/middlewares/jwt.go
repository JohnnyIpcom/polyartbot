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
			restErr := utils.NewUnauthorizedError("No Authorization token in header", nil)
			c.AbortWithStatusJSON(restErr.Status(), restErr)
			return
		}

		tokenString := authHeader[len(BEARER_SCHEMA):]

		token, err := j.ValidateToken(tokenString)
		if !token.Valid {
			restErr := utils.NewUnauthorizedError("Unathorized!", err)
			c.AbortWithStatusJSON(restErr.Status(), restErr)
		}
	}
}
