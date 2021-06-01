package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/johnnyipcom/polyartbot/cdn/models"
	"github.com/johnnyipcom/polyartbot/cdn/services"
	"github.com/johnnyipcom/polyartbot/cdn/utils"
)

type LoginController interface {
	Login(c *gin.Context)
}

type loginController struct {
	loginService services.LoginService
	jWtService   services.JWTService
}

func NewLoginController(l services.LoginService, j services.JWTService) LoginController {
	return &loginController{
		loginService: l,
		jWtService:   j,
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

	token, err := l.jWtService.GenerateToken(credentials.Username, true)
	if err != nil {
		restErr := utils.NewInternalServerError("JWT token generation error", err)
		c.JSON(restErr.Status(), restErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
	})
}
