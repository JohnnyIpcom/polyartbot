package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthController interface {
	Health(c *gin.Context)
}

type healthController struct{}

var _ HealthController = &healthController{}

func NewHealthController() HealthController {
	return &healthController{}
}

func (h *healthController) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "OK",
	})
}
