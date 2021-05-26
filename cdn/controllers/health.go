package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Health interface {
	Health(c *gin.Context)
}

type health struct{}

var _ Health = &health{}

func NewHealthController() Health {
	return &health{}
}

func (h *health) Health(c *gin.Context) {
	c.String(http.StatusOK, "OK")
}
