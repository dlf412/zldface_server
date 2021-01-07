package recognition

import (
	"os"
)

var appid, key = os.Getenv("ARCSOFT_FACE_APPID"), os.Getenv("ARCSOFT_FACE_KEY")

func BGR24Data(image interface{}) (width, height int, data []uint8, err error) {

	img, err := DecodeImage(image)
	if err != nil {
		return width, height, data, err
	}
	height = GetImageHeight(img)
	width = GetImageWidth(img)
	width = width - width%4

	imgMatrix, err := ResizeForMatrix(image, width, height)

	if err != nil {
		return
	}
	for starty := 0; starty < height; starty++ {
		for startx := 0; startx < width; startx++ {
			R := imgMatrix[starty][startx][0]
			G := imgMatrix[starty][startx][1]
			B := imgMatrix[starty][startx][2]
			data = append(data, B, G, R)
		}
	}
	return
}

type FaceImage struct {
	Width     int
	Height    int
	ImageData []uint8
	SingleFaceInfo
	FaceFeature
}

type Engine struct {
	*FaceEngine
}

func NewEngine() (*Engine, error) {
	engine := Engine{}
	// 初始化引擎
	fe, err := NewFaceEngine(DetectModeImage,
		OrientPriority0,
		12,
		1,
		EnableFaceDetect|EnableFaceRecognition)

	if err.(EngineError).Code == 90115 {
		if err := Activation(appid, key); err != nil {
			return nil, err
		}
		fe, err = NewFaceEngine(DetectModeImage,
			OrientPriority0,
			12,
			1,
			EnableFaceDetect|EnableFaceRecognition)
	}

	if err != nil {
		return nil, err
	}

	engine.FaceEngine = fe
	return &engine, nil
}

func (e *Engine) DetectFace(img interface{}) (*FaceImage, error) {
	face := FaceImage{}
	width, height, imagedata, err := BGR24Data(img)
	if err != nil {
		return nil, err
	}
	face.Width = width
	face.Height = height
	face.ImageData = imagedata
	info, err := e.DetectFaces(face.Width, face.Height, ColorFormatBGR24, face.ImageData)
	if err != nil {
		return nil, err
	}
	singleFaceInfoArr := GetSingleFaceInfo(info)
	if len(singleFaceInfoArr) == 0 {
		return nil, err
	}
	face.SingleFaceInfo = singleFaceInfoArr[0]
	return &face, nil
}

func (e *Engine) ExtractFace(image *FaceImage) (FaceFeature, error) {
	return e.FaceFeatureExtract(image.Width, image.Height, ColorFormatBGR24, image.ImageData, image.SingleFaceInfo)
}

func (e *Engine) CompareFace(f1, f2 FaceFeature) (float32, error) {
	return e.FaceFeatureCompare(f1, f2)
}

func (e *Engine) Destory() error {
	return e.Destory()
}

func CompareFeature(arr1, arr2 []byte) (score float32, err error) {
	var eng *Engine
	eng, err = NewEngine()
	if err != nil {
		return
	}
	defer eng.Destory()
	return eng.FaceFeatureCompareEx(arr1, arr2)
}

func (e *Engine) CompareImgFeature(img interface{}, arr []byte) (score float32, err error) {
	var eng *Engine
	eng, err = NewEngine()
	if err != nil {
		return
	}
	defer eng.Destory()

	var face *FaceImage
	face, err = eng.DetectFace(img)
	if err != nil {
		return
	}

	var f FaceFeature
	f, err = eng.ExtractFace(face)
	if err != nil {
		return
	}
	defer f.Release()
	return eng.FaceFeatureCompareEx(f.Feature, arr)

}

func CompareImg(img1, img2 interface{}) (score float32, err error) {
	var eng *Engine
	eng, err = NewEngine()

	if err != nil {
		return
	}
	defer eng.Destory()
	var face1, face2 *FaceImage
	face1, err = eng.DetectFace(img1)
	if err != nil {
		return
	}
	face2, err = eng.DetectFace(img2)
	if err != nil {
		return
	}

	var f1, f2 FaceFeature

	f1, err = eng.ExtractFace(face1)
	if err != nil {
		return
	}
	defer f1.Release()

	f2, err = eng.ExtractFace(face2)
	if err != nil {
		return
	}
	defer f2.Release()
	return eng.CompareFace(f1, f2)
}
