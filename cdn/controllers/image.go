package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/johnnyipcom/polyartbot/cdn/services"
	"github.com/johnnyipcom/polyartbot/glue"
)

type ImageController interface {
	Post(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
}

type imageController struct {
	image services.ImageService
}

var _ ImageController = &imageController{}

func NewImageController(image services.ImageService) ImageController {
	return &imageController{
		image: image,
	}
}

func (i *imageController) Post(c *gin.Context) {
	var results []glue.RespFile

	form, _ := c.MultipartForm()
	headers := form.File["file"]

	for _, header := range headers {
		file, _ := header.Open()
		fileID, size, err := i.image.Upload(file, *header)
		if err != nil {
			restErr := glue.NewBadRequestError(err.Error())
			c.JSON(restErr.Status(), restErr)
			return
		}

		if err := i.image.Publish(fileID); err != nil {
			restErr := glue.NewInternalServerError("Can't post ID to RabbitMQ", err)
			c.JSON(restErr.Status(), restErr)
			return
		}

		results = append(results, glue.NewFileResponse(fileID, size))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Files uploaded successfully",
		"files":   results,
	})
}

func (i *imageController) Get(c *gin.Context) {
	data, err := i.image.Download(c.Param("filename"))
	if err != nil {
		restErr := glue.NewBadRequestError(err.Error())
		c.JSON(restErr.Status(), restErr)
		return
	}

	respGet := glue.NewFileGet("File returned successfully", data)
	c.JSON(http.StatusOK, respGet)
}

func (i *imageController) Delete(c *gin.Context) {
	err := i.image.Delete(c.Param("filename"))
	if err != nil {
		restErr := glue.NewBadRequestError(err.Error())
		c.JSON(restErr.Status(), restErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}
