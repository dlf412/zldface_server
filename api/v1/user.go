package v1

import (
	"bytes"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"zldface_server/model"
	"zldface_server/recognition"
	//"go.uber.org/zap"
	"net/http"
	"zldface_server/config"
	"zldface_server/model/request"
)

//  godoc
// @Summary Create a user
// @Description create a user with faceFile or faceFeature.
// @Accept  multipart/form-data
// @Produce  json
// @param uid formData string true "user id"
// @param name formData string true "name"
// @Param faceFile formData file false "faceFile文件"
// @param gid formData string false "group id"
// @param faceFeature formData file false "人脸特征文件, binary格式"
// @param FaceImagePath formData string false "人脸照片路径（服务器已存在的相对路径）"
// @Success 201 {object} model.FaceUser
// @Router /users/v1 [post]
func CreateUser(c *gin.Context) {
	var U request.FaceUser
	if err := c.Bind(&U); err != nil {
		config.Logger.Info(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	user := model.FaceUser{
		Uid:           U.Uid,
		Name:          U.Name,
		FaceFeature:   nil,
		FaceImagePath: U.FaceImagePath,
		Groups:        nil,
	}

	if len(U.Gid) > 0 {
		groups := []model.FaceGroup{}
		config.DB.Where("Gid in ?", U.Gid).Find(&groups)
		user.Groups = groups
	}

	if U.FaceFile != nil {
		faceFile, err := U.FaceFile.Open()
		if err != nil {
			config.Logger.Error("打开文件失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			return
		}

		defer faceFile.Close()

		user.FaceImagePath = U.Uid + U.FaceFile.Filename
		user.FaceFeature, err = recognition.FeatureByteArr(faceFile)
		if err != nil {
			config.Logger.Warn("图片提取人脸特征失败", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			return
		} else {
			go func(f string) {
				c.SaveUploadedFile(U.FaceFile, config.RegDir+"/"+f)
			}(user.FaceImagePath)
		}
	} else {
		if U.FaceFeature != nil && U.FaceFeature.Size == 1032 {
			ff, err := U.FaceFeature.Open()
			if err != nil {
				config.Logger.Error("读取FaceFeature数据失败", zap.Error(err))
				c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
				return
			}
			defer ff.Close()
			user.FaceFeature = make([]byte, 1032)
			io.Copy(bytes.NewBuffer(user.FaceFeature), ff)
		} else {
			var err error
			user.FaceFeature, err = recognition.FeatureByteArr(config.RegDir + "/" + user.FaceImagePath)
			if err != nil {
				config.Logger.Warn("图片提取人脸特征失败", zap.Error(err))
				c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
				return
			}
		}
	}
	if err := config.DB.Create(&user).Error; err != nil {
		config.Logger.Error("保存数据失败", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

//  godoc
// @Summary Match a user
// @Description post a faceFile to match a user in a group.
// @Accept  multipart/form-data
// @Param faceFile formData file true "faceFile"
// @param gid formData string true "group id"
// @Produce  json
// @Success 201 {object} recognition.Closest
// @Router /user/match/v1 [post]
func MatchUser(c *gin.Context) {
	var M request.FaceUserMatch
	if err := c.Bind(&M); err != nil {
		config.Logger.Info(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}

	// 加载gourp对应的人脸库
	group := new(model.FaceGroup)
	config.DB.Where("`Gid`=?", M.Gid).First(group)
	if group.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"err": "group不存在"})
		return
	}
	eng, err := recognition.NewEngine()
	if err != nil {
		config.Logger.Error(err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"err": "人脸识别引擎初始化失败"})
		return
	}
	defer eng.Destroy()

	ff, _ := M.FaceFile.Open()
	defer ff.Close()

	matches, err := eng.SearchN(ff, group.FaceFeatures(), 1, 0.8)
	if err != nil {
		config.Logger.Error("人脸查找发生错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"err": "人脸匹配发生错误"})
		return
	}
	if matches == nil || len(matches) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"err": "未匹配到任何人脸"})
		return
	}
	config.Logger.Info("匹配成功", zap.Any("结果", matches[0]))
	c.JSON(http.StatusOK, matches[0])
}
