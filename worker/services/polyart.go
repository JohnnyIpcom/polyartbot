package services

import (
	"bytes"
	"image"
	_ "image/gif"
	"image/jpeg"
	_ "image/png"
	"runtime"

	"github.com/fogleman/primitive/primitive"
	"github.com/johnnyipcom/polyartbot/utils"
	"github.com/johnnyipcom/polyartbot/worker/config"
	"go.uber.org/zap"
)

type PolyartService interface {
	JustCopy(data []byte) ([]byte, error)
	Convert(data []byte) ([]byte, error)
}

type polyartService struct {
	cfg config.Polyart
	log *zap.Logger
}

func NewPolyartService(cfg config.Config, log *zap.Logger) PolyartService {
	return &polyartService{
		cfg: cfg.Polyart,
		log: log.Named("polyartService"),
	}
}

func (p *polyartService) JustCopy(data []byte) ([]byte, error) {
	result := make([]byte, len(data))
	copy(result, data)

	return result, nil
}

func (p *polyartService) Convert(data []byte) ([]byte, error) {
	image, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	color := primitive.MakeColor(primitive.AverageImageColor(image))
	_, sz := utils.MinMax(image.Bounds().Dx(), image.Bounds().Dy())
	model := primitive.NewModel(image, color, sz, runtime.NumCPU())

	for i := 0; i < p.cfg.Steps; i++ {
		model.Step(primitive.ShapeType(p.cfg.Shape), 128, 0)
	}

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(
		buffer,
		model.Context.Image(),
		&jpeg.Options{
			Quality: 95,
		},
	); err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
