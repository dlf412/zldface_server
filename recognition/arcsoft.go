package recognition

import (
	"errors"
	"os"
	"sort"
	"sync"
	"time"
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

	imgMatrix, err := ResizeForMatrix(img, width, height)

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
}

type Engine struct {
	*FaceEngine
}

type Closest struct {
	Key   interface{} // 用户自定义的key, 可以是身份证号，可以是文件路径等
	Score float32
}

func NewEngine() (*Engine, error) {
	engine := Engine{}
	// 初始化引擎
	fe, err := NewFaceEngine(DetectModeImage,
		OrientPriority0,
		12,
		1,
		EnableFaceDetect|EnableFaceRecognition)

	if err != nil && err.(EngineError).Code == 90115 {
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
		return nil, errors.New("未检测到人脸")
	}
	face.SingleFaceInfo = singleFaceInfoArr[0]
	return &face, nil
}

func (e *Engine) ExtractFeature(image *FaceImage) (FaceFeature, error) {
	return e.FaceFeatureExtract(image.Width, image.Height, ColorFormatBGR24, image.ImageData, image.SingleFaceInfo)
}

func (e *Engine) ExtractFeatureByteArr(image *FaceImage) ([]byte, error) {
	feature, err := e.FaceFeatureExtract(image.Width, image.Height, ColorFormatBGR24, image.ImageData, image.SingleFaceInfo)
	if err != nil {
		return nil, err
	}
	defer feature.Release()
	return feature.Feature, nil
}

func (e *Engine) Img2FeatureByteArr(img interface{}) ([]byte, error) {
	face, err := e.DetectFace(img)
	if err != nil {
		return nil, err
	}
	return e.ExtractFeatureByteArr(face)
}

func (e *Engine) CompareFeature(f1, f2 interface{}) (score float32, err error) {
	// 支持图片路径, io.reader, []byte, FaceFeature
	var feature1, feature2 []byte
	switch f1.(type) {
	case []byte:
		feature1 = f1.([]byte)
	case FaceFeature:
		feature1 = f1.(FaceFeature).Feature
	default:
		face, err := e.DetectFace(f1)
		if err != nil {
			return 0.0, err
		}
		feature1, err = e.ExtractFeatureByteArr(face)
		if err != nil {
			return 0.0, err
		}
	}

	switch f2.(type) {
	case []byte:
		feature2 = f2.([]byte)
	case FaceFeature:
		feature2 = f2.(FaceFeature).Feature
	default:
		face, err := e.DetectFace(f2)
		if err != nil {
			return 0.0, err
		}
		feature2, err = e.ExtractFeatureByteArr(face)
		if err != nil {
			return 0.0, err
		}
	}

	return e.FaceFeatureCompareEx(feature1, feature2)
}

func (e *Engine) SearchN(f1 interface{}, byteFeatures map[interface{}][]byte, top int, threshold float32) ([]Closest, error) {
	var feature1 []byte
	switch f1.(type) {
	case []byte:
		feature1 = f1.([]byte)
	case FaceFeature:
		feature1 = f1.(FaceFeature).Feature
	default:
		face, err := e.DetectFace(f1)
		if err != nil {
			return nil, err
		}
		feature1, err = e.ExtractFeatureByteArr(face)
		if err != nil {
			return nil, err
		}
	}
	t_cnt := len(byteFeatures)
	if t_cnt == 0 {
		return nil, errors.New("byteFeatures is Empty")
	}
	// 创建任务通道
	tasks := make(chan map[interface{}][]byte, t_cnt)
	results := make(chan Closest, t_cnt)
	max_groutine := 100

	if t_cnt < max_groutine {
		max_groutine = t_cnt
	}
	wg := sync.WaitGroup{}
	wg.Add(max_groutine)
	for gr := 1; gr <= max_groutine; gr++ { //
		go func() {
			for {
				select {
				case t, ok := <-tasks:
					if !ok {
						goto END
					}
					for k, v := range t {
						score, _ := e.FaceFeatureCompareEx(feature1, v)
						results <- Closest{Key: k, Score: score}
					}
				}
			}
		END:
			wg.Done()
		}()
	}

	// 通道发送任务
	for k, v := range byteFeatures {
		tasks <- map[interface{}][]byte{k: v}
	}
	// 通道接收结果
	res := []Closest{}
FOR:
	for i := 0; i < t_cnt; i++ {
		select {
		case r := <-results:
			if r.Score >= threshold {
				res = append(res, r)
			}
		case <-time.After(time.Second * 5): // 5秒超时,强制退出
			break FOR
		}
	}
	close(tasks)
	// 等待所有task完成
	wg.Wait()
	// 安全关闭results
	close(results)
	sort.Slice(res, func(i, j int) bool { return res[i].Score > res[j].Score })
	if top < len(res) {
		return res[0:top], nil
	} else {
		return res, nil
	}

}

func FeatureByteArr(img interface{}) (feature []byte, err error) {
	var eng *Engine
	eng, err = NewEngine()
	if err != nil {
		return
	}
	defer eng.Destroy()
	var face *FaceImage
	face, err = eng.DetectFace(img)
	if err != nil {
		return
	}
	return eng.ExtractFeatureByteArr(face)
}

//func CompareFeature(arr1, arr2 []byte) (score float32, err error) {
//	var eng *Engine
//	eng, err = NewEngine()
//	if err != nil {
//		return
//	}
//	defer eng.Destroy()
//
//	return eng.FaceFeatureCompareEx(arr1, arr2)
//}

//func CompareImgFeature(img interface{}, arr []byte) (score float32, err error) {
//	var eng *Engine
//	eng, err = NewEngine()
//	if err != nil {
//		return
//	}
//	defer eng.Destroy()
//
//	var face *FaceImage
//	face, err = eng.DetectFace(img)
//	if err != nil {
//		return
//	}
//
//	var f FaceFeature
//	f, err = eng.ExtractFeature(face)
//	if err != nil {
//		return
//	}
//	defer f.Release()
//	return eng.FaceFeatureCompareEx(f.Feature, arr)
//
//}
//
//func CompareImg(img1, img2 interface{}) (score float32, err error) {
//	var eng *Engine
//	eng, err = NewEngine()
//
//	if err != nil {
//		return
//	}
//	defer eng.Destroy()
//	var face1, face2 *FaceImage
//	face1, err = eng.DetectFace(img1)
//	if err != nil {
//		return
//	}
//	face2, err = eng.DetectFace(img2)
//	if err != nil {
//		return
//	}
//
//	var f1, f2 FaceFeature
//
//	f1, err = eng.ExtractFeature(face1)
//	if err != nil {
//		return
//	}
//	defer f1.Release()
//
//	f2, err = eng.ExtractFeature(face2)
//	if err != nil {
//		return
//	}
//	defer f2.Release()
//	return eng.CompareFeature(f1, f2)
//}