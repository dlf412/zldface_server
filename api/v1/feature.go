package v1

import (
	"encoding/base64"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"path"
	"zldface_server/cache"
	"zldface_server/config"
	"zldface_server/model/request"
	"zldface_server/recognition"
	"zldface_server/utils"
)

type ExtractFeatureRes struct {
	FaceFeature   []byte `json:"faceFeature"`
	FaceImagePath string `json:"faceImagePath"`
}

type CompareFileRes struct {
	FaceImagePath1 string  `json:"faceImagePath1"` // 人脸路径1
	FaceImagePath2 string  `json:"faceImagePath2"` // 人脸路径2
	Score          float32 `json:"score"`
}

func saveFileAndExtractFeature(faceFile io.ReadSeeker, mustBeFace bool) (string, []byte, error) {
	faceImagePath := utils.MD5RelativePath(faceFile)
	faceFeature, err := recognition.FeatureByteArr(faceFile)
	if err != nil && mustBeFace {
		return "", nil, err
	} else {
		faceFile.Seek(0, io.SeekStart) // seek到0
		if err := utils.SaveFile(faceFile, path.Join(config.RegDir, faceImagePath)); err != nil {
			return "", nil, err
		} else {
			return faceImagePath, faceFeature, nil
		}
	}
}

//  godoc
// @Summary save a FaceImage and extract faceFeature
// @Description post a face image then return the save path and faceFeature.
// @Accept  multipart/form-data
// @Param faceFile formData file true "faceFile"
// @Produce  json
// @Success 201 {object} ExtractFeatureRes
// @Router /faceImage/v1 [post]
func SaveFaceImage(c *gin.Context) {
	file, err := c.FormFile("faceFile")
	if err != nil {
		config.Logger.Info(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	faceFile, err := file.Open()
	if err != nil {
		config.Logger.Error("打开文件失败", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
		return
	}

	defer faceFile.Close()
	// 相对路径
	faceImagePath, faceFeature, err := saveFileAndExtractFeature(faceFile, true)
	if err != nil {
		config.Logger.Warn("图片提取人脸特征失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	} else {
		cache.UpdatePathFeature(faceImagePath, faceFeature)
		c.JSON(http.StatusCreated, ExtractFeatureRes{FaceImagePath: faceImagePath})
	}
}

//  godoc
// @Summary get faceFeature by faceImagePath
// @Description faceImagePath return from /faceImage/v1
// @Produce  json
// @Param faceImagePath query string true "faceImagePath"
// @Success 200 {object} ExtractFeatureRes
// @Router /faceFeature/v1 [get]
func GetFaceFeature(c *gin.Context) {
	faceImagePath := c.Query("faceImagePath")
	if faceImagePath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"err": "faceImagePath is empty error"})
		return
	}
	feature, err := cache.GetPathFeature(faceImagePath)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	c.JSON(http.StatusOK, ExtractFeatureRes{FaceImagePath: faceImagePath, FaceFeature: feature})
}

//  godoc
// @Summary compare two faceFeature
// @Description compare two faceFeature and return the score (0<score<=1)
// @Produce json
// @Accept json
// @Param data body request.FaceFeatures true "人脸特征1, 人脸特征2"
// @Success 200 {string} string "{"score":0.90}"
// @Router /featureCompare/v1 [post]
func CompareFaceFeature(c *gin.Context) {
	var features request.FaceFeatures
	if err := c.ShouldBind(&features); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	f1, _ := base64.StdEncoding.DecodeString(features.Feature1)
	f2, _ := base64.StdEncoding.DecodeString(features.Feature2)
	score, err := recognition.CompareFeature(f1, f2)
	if err != nil {
		config.Logger.Warn("比对特征失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"socre": score})
}

// godoc
// @Summary compare two faceImage files
// @Description compare two faceImage files and return the score and save path(0<score<=1)
// @Accept  multipart/form-data
// @Produce json
// @Param face1 formData string false "人脸照片1路径或者特征，有此参数忽略人脸照片1"
// @Param face2 formData string false "人脸照片2路径或者特征，有此参数忽略人脸照片2"
// @Param faceFile1 formData file false "人脸照片1"
// @Param faceFile2 formData file false "人脸照片2"
// @Success 200 {object} CompareFileRes
// @Router /faceCompare/v1 [post]
func CompareFaceFile(c *gin.Context) {
	var faces request.FaceFiles
	if err := c.ShouldBind(&faces); err != nil {
		config.Logger.Info("参数错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	path1 := faces.Face1
	path2 := faces.Face2
	var f1, f2 []byte
	var err error

	if path1 == "" {
		file, _ := faces.FaceFile1.Open()
		defer file.Close()
		path1, f1, err = saveFileAndExtractFeature(file, false) // 图一可以忽略
	} else {
		if len(path1) == 1376 {
			f1, err = base64.StdEncoding.DecodeString(path1)
		} else {
			f1, err = cache.GetPathFeature(path1)
		}
		path1 = ""
	}
	if err != nil {
		config.Logger.Warn("保存图片1或者提取人脸特征1失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	if path2 == "" {
		file, _ := faces.FaceFile2.Open()
		defer file.Close()
		path2, f2, err = saveFileAndExtractFeature(file, true) // 图二必须是人脸
	} else {
		if len(path2) == 1376 {
			f2, err = base64.StdEncoding.DecodeString(path2)
		} else {
			f2, err = cache.GetPathFeature(path2)
		}
		path2 = ""
	}
	if err != nil {
		config.Logger.Warn("提取特征2失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	var score float32
	if f1 != nil && f2 != nil {
		score, err = recognition.CompareFeature(f1, f2)
		if err != nil {
			score = 0.0
			config.Logger.Warn("比对特征失败", zap.Error(err))
		}
	} else {
		score = 0.0
	}
	//if path1 != "" {
	//	cache.UpdatePathFeature(path1, f1)
	//}
	if path2 != "" {
		cache.UpdatePathFeature(path2, f2)
	}
	c.JSON(http.StatusOK, CompareFileRes{
		FaceImagePath1: path1,
		FaceImagePath2: path2,
		Score:          score,
	})
}
