package controllers

import (
	"mime/multipart"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/johnnyipcom/polyartbot/cdn/services"
	"github.com/johnnyipcom/polyartbot/models"
	"golang.org/x/sync/errgroup"
)

type ImageController interface {
	Post(c *gin.Context)
	Get(c *gin.Context)
	Delete(c *gin.Context)
}

type imageController struct {
	image    services.ImageService
	rabbitMQ services.RabbitMQService
}

var _ ImageController = &imageController{}

func NewImageController(i services.ImageService, r services.RabbitMQService) ImageController {
	return &imageController{
		image:    i,
		rabbitMQ: r,
	}
}

func (i *imageController) Post(c *gin.Context) {
	type imageStruct struct {
		From int64 `form:"from"`
		To   int64 `form:"to"`
	}

	var image imageStruct
	if err := c.BindQuery(&image); err != nil {
		restErr := models.NewBadRequestError(err.Error())
		c.JSON(restErr.Status(), restErr)
		return
	}

	form, _ := c.MultipartForm()
	headers := form.File["file"]

	var lenMap sync.Map
	g := errgroup.Group{}
	for _, header := range headers {
		func(header *multipart.FileHeader) {
			uuid := uuid.New()
			g.Go(func() error {
				file, err := header.Open()
				if err != nil {
					return nil
				}

				metadata := make(map[string]string)
				metadata["from"] = strconv.FormatInt(image.From, 10)
				metadata["to"] = strconv.FormatInt(image.To, 10)

				len, err := i.image.Upload(uuid.String(), file, *header, metadata)
				if err != nil {
					return nil
				}

				lenMap.Store(uuid.String(), len)
				return nil
			})
		}(header)
	}

	if err := g.Wait(); err != nil {
		restErr := models.NewInternalServerError("internal storage error", err)
		c.JSON(restErr.Status(), restErr)
		return
	}

	if image.From != 0 {
		lenMap.Range(func(key interface{}, value interface{}) bool {
			fileID := key.(string)
			g.Go(func() error {
				return i.rabbitMQ.Publish(models.NewRabbitMQImage(fileID, image.From))
			})
			return true
		})
	}

	if err := g.Wait(); err != nil {
		restErr := models.NewInternalServerError("internal rabbitMQ error", err)
		c.JSON(restErr.Status(), restErr)
		return
	}

	results := make([]models.RespFile, 0)
	lenMap.Range(func(key interface{}, value interface{}) bool {
		fileID := key.(string)
		length := value.(int)
		results = append(results, models.NewFileResponse(fileID, length))
		return true
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Files uploaded successfully",
		"files":   results,
	})
}

func (i *imageController) Get(c *gin.Context) {
	fileID := c.Param("filename")

	var (
		data     []byte
		metadata map[string]string
	)

	g := errgroup.Group{}
	g.Go(func() error {
		d, err := i.image.Download(fileID)
		if err != nil {
			return err
		}

		data = d
		return nil
	})

	g.Go(func() error {
		m, err := i.image.GetMetadata(fileID)
		if err != nil {
			return err
		}

		metadata = m
		return nil
	})

	if err := g.Wait(); err != nil {
		restErr := models.NewBadRequestError(err.Error())
		c.JSON(restErr.Status(), restErr)
		return
	}

	respGet := models.NewFileGet("File returned successfully", data, metadata)
	c.JSON(http.StatusOK, respGet)
}

func (i *imageController) Delete(c *gin.Context) {
	err := i.image.Delete(c.Param("filename"))
	if err != nil {
		restErr := models.NewBadRequestError(err.Error())
		c.JSON(restErr.Status(), restErr)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File deleted successfully",
	})
}
