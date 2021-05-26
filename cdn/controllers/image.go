package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/johnnyipcom/polyartbot/cdn/storage"
	"github.com/johnnyipcom/polyartbot/cdn/utils"
)

type Image interface {
	Upload(c *gin.Context)
	UploadMany(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
}

type image struct {
	storage storage.Storage
}

var _ Image = &image{}

func NewImageController(storage storage.Storage) Image {
	return &image{
		storage: storage,
	}
}

func (i *image) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		restErr := utils.NewBadRequestError(err.Error())
		c.JSON(restErr.Status(), restErr)
		return
	}
	defer file.Close()

	name, size, err := i.storage.Upload(file, *header)
	if err != nil {
		restErr := utils.NewBadRequestError(err.Error())
		c.JSON(restErr.Status(), restErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File uploaded successfully",
		"file":    utils.NewFileResponse(name, size),
	})
}

func (i *image) UploadMany(c *gin.Context) {
	var results []utils.RespFile

	form, _ := c.MultipartForm()
	headers := form.File["files"]

	for _, header := range headers {
		file, _ := header.Open()
		name, size, err := i.storage.Upload(file, *header)
		if err != nil {
			restErr := utils.NewBadRequestError(err.Error())
			c.JSON(restErr.Status(), restErr)
			return
		}
		results = append(results, utils.NewFileResponse(name, size))
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Files uploaded successfully",
		"files":   results,
	})
}

func (i *image) Get(c *gin.Context) {
	data, err := i.storage.Download(c.Param("filename"))
	if err != nil {
		restErr := utils.NewBadRequestError(err.Error())
		c.JSON(restErr.Status(), restErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File returned successfully",
		"data":    data,
	})
}

func (i *image) Delete(c *gin.Context) {
	err := i.storage.Delete(c.Param("filename"))
	if err != nil {
		restErr := utils.NewBadRequestError(err.Error())
		c.JSON(restErr.Status(), restErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}
