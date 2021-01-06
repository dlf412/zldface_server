package arcsoft

import (
	"fmt"
	"github.com/windosx/face-engine/v3"
	"github.com/windosx/face-engine/v3/util"
)

var appid, key = "my appid", "my key"

type BGR24Imager interface {
	Width() int
	Height() int
	BGRData() []uint8
}

type FaceImage struct {
	Width     int
	Height    int
	ImageData []uint8
	face_engine.SingleFaceInfo
	face_engine.FaceFeature
}

type Engine struct {
	*face_engine.FaceEngine
}

var width, height = util.GetImageWidthAndHeight("./test.jpg")
var imageData = util.GetResizedBGR("./test.jpg")

func New() (*Engine, error) {
	engine := Engine{}
	// 初始化引擎
	fe, err := face_engine.NewFaceEngine(face_engine.DetectModeImage,
		face_engine.OrientPriority0,
		12,
		1,
		face_engine.EnableFaceDetect|face_engine.EnableFaceRecognition)

	if err.(face_engine.EngineError).Code == 90115 {
		if err := face_engine.Activation(appid, key); err != nil {
			return nil, err
		}
		fe, err = face_engine.NewFaceEngine(face_engine.DetectModeImage,
			face_engine.OrientPriority0,
			12,
			50,
			face_engine.EnableFaceDetect|face_engine.EnableFaceRecognition)
	}

	if err != nil {
		return nil, err
	}

	engine.FaceEngine = fe
	return &engine, nil
}

func (e *Engine) DetectFace(img BGR24Imager) (*FaceImage, error) {
	face := FaceImage{}
	face.Width = img.Width()
	face.Width -= face.Width % 4
	face.Height = img.Height()
	face.ImageData = img.BGRData()

	info, err := e.DetectFaces(face.Width, face.Height, face_engine.ColorFormatBGR24, img.BGRData())
	if err != nil {
		return nil, err
	}
	singleFaceInfoArr := face_engine.GetSingleFaceInfo(info)
	if len(singleFaceInfoArr) == 0 {
		return nil, err
	}
	face.SingleFaceInfo = singleFaceInfoArr[0]
	return &face, nil
}

func (e *Engine) ExtractFace(image *FaceImage) (face_engine.FaceFeature, error) {
	return e.FaceFeatureExtract(image.Width, image.Height, face_engine.ColorFormatBGR24, image.ImageData, image.SingleFaceInfo)
}

func (e *Engine) CompareFace(f1, f2 face_engine.FaceFeature) (float32, error) {
	return e.FaceFeatureCompare(f1, f2)
}

func (e *Engine) Destory() error {
	return e.Destory()
}
