package controllers

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/johnnyipcom/polyartbot/cdn/models"
	"github.com/johnnyipcom/polyartbot/cdn/services"
	"github.com/johnnyipcom/polyartbot/cdn/utils"
)

type LoginController interface {
	Login(c *gin.Context)
	Logout(c *gin.Context)
	Refresh(c *gin.Context)
}

type loginController struct {
	loginService services.LoginService
	jwtService   services.JWTService
	cacheService services.CacheService
}

func NewLoginController(l services.LoginService, j services.JWTService, c services.CacheService) LoginController {
	return &loginController{
		loginService: l,
		jwtService:   j,
		cacheService: c,
	}
}

func (l *loginController) Login(c *gin.Context) {
	var credentials models.Credentials
	err := c.ShouldBind(&credentials)
	if err != nil {
		restErr := utils.NewBadRequestError(err.Error())
		c.JSON(restErr.Status(), restErr)
		return
	}

	isAuthenticated := l.loginService.Login(credentials.Username, credentials.Password)
	if !isAuthenticated {
		restErr := utils.NewUnauthorizedError("Invalid username or password", nil)
		c.JSON(restErr.Status(), restErr)
		return
	}

	accessToken, refreshToken, err := l.jwtService.GenerateTokens(credentials.Username, true)
	if err != nil {
		restErr := utils.NewInternalServerError("JWT token generation error", err)
		c.JSON(restErr.Status(), restErr)
		return
	}

	if err := l.cacheService.SaveAuth(credentials.Username, accessToken, refreshToken); err != nil {
		restErr := utils.NewInternalServerError("JWT token storage error", err)
		c.JSON(restErr.Status(), restErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "login success",
		"access_token":  accessToken.Token,
		"refresh_token": refreshToken.Token,
	})
}

func (l *loginController) Logout(c *gin.Context) {
	const BEARER_SCHEMA = "Bearer "
	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, BEARER_SCHEMA) {
		restErr := utils.NewUnauthorizedError("no valid authorization bearer", nil)
		c.AbortWithStatusJSON(restErr.Status(), restErr)
	}

	token := authHeader[len(BEARER_SCHEMA):]
	uuid, username, err := l.jwtService.ExtractTokenMetadata(token, true)
	if err != nil {
		restErr := utils.NewUnauthorizedError("unauthorized", err)
		c.JSON(restErr.Status(), restErr)
		return
	}

	if !l.loginService.Logout(username) {
		restErr := utils.NewUnauthorizedError("logout failed", nil)
		c.JSON(restErr.Status(), restErr)
		return
	}

	if err := l.cacheService.DeleteAuth(uuid); err != nil {
		restErr := utils.NewUnauthorizedError("unauthorized", err)
		c.JSON(restErr.Status(), restErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "logout success",
	})
}

func (l *loginController) Refresh(c *gin.Context) {
	mapToken := map[string]string{}
	if err := c.ShouldBindJSON(&mapToken); err != nil {
		restErr := utils.NewBadRequestError(err.Error())
		c.JSON(restErr.Status(), restErr)
		return
	}

	refreshTokenString := mapToken["refresh_token"]
	uuid, username, err := l.jwtService.ExtractTokenMetadata(refreshTokenString, true)
	if err != nil {
		restErr := utils.NewUnauthorizedError("refresh token invalid or expired", err)
		c.JSON(restErr.Status(), restErr)
		return
	}

	if err := l.cacheService.DeleteAuth(uuid); err != nil {
		restErr := utils.NewInternalServerError("JWT token storage error", err)
		c.JSON(restErr.Status(), restErr)
		return
	}

	accessToken, refreshToken, err := l.jwtService.GenerateTokens(username, true)
	if err != nil {
		restErr := utils.NewInternalServerError("JWT token generation error", err)
		c.JSON(restErr.Status(), restErr)
		return
	}

	if err := l.cacheService.SaveAuth(username, accessToken, refreshToken); err != nil {
		restErr := utils.NewInternalServerError("JWT token storage error", err)
		c.JSON(restErr.Status(), restErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "refresh token success",
		"access_token":  accessToken.Token,
		"refresh_token": refreshToken.Token,
	})
}
