package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/johnnyipcom/polyartbot/cdn/services"
	"github.com/johnnyipcom/polyartbot/cdn/utils"
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
	var results []utils.RespFile

	userFrom, userTo := c.Query("from"), c.Query("to")
	if userFrom == "" && userTo == "" {
		restErr := utils.NewBadRequestError("'from' or 'to' should either be non-nil")
		c.JSON(restErr.Status(), restErr)
		return
	}

	form, _ := c.MultipartForm()
	headers := form.File["file"]

	for _, header := range headers {
		file, _ := header.Open()
		fileID, size, err := i.image.Upload(file, *header)
		if err != nil {
			restErr := utils.NewBadRequestError(err.Error())
			c.JSON(restErr.Status(), restErr)
			return
		}

		if err := i.image.Publish(fileID); err != nil {
			restErr := utils.NewInternalServerError("Can't post ID to RabbitMQ", err)
			c.JSON(restErr.Status(), restErr)
			return
		}

		results = append(results, utils.NewFileResponse(fileID, size))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Files uploaded successfully",
		"files":   results,
	})
}

func (i *imageController) Get(c *gin.Context) {
	fileID := c.Param("filename")

	data, err := i.image.Download(fileID)
	if err != nil {
		restErr := utils.NewBadRequestError(err.Error())
		c.JSON(restErr.Status(), restErr)
		return
	}

	metadata, err := i.image.GetMetadata(fileID)
	if err != nil {
		restErr := utils.NewInternalServerError("Metadata error", err)
		c.JSON(restErr.Status(), restErr)
		return
	}

	respGet := utils.NewFileGet("File returned successfully", data, metadata)
	c.JSON(http.StatusOK, respGet)
}

func (i *imageController) Delete(c *gin.Context) {
	err := i.image.Delete(c.Param("filename"))
	if err != nil {
		restErr := utils.NewBadRequestError(err.Error())
		c.JSON(restErr.Status(), restErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}
