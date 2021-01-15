package v1

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"path"
	"sync"
	"zldface_server/cache"
	"zldface_server/config"
	"zldface_server/model"
	"zldface_server/model/request"
	"zldface_server/model/response"
	"zldface_server/recognition"
	"zldface_server/utils"
)

//  godoc
// @Summary get user by uid
// @Description get user by uid if
// @Produce  json
// @param uid path string true "user id"
// @Success 200 {object} model.FaceUser
// @Router /users/v1/{uid} [get]
func GetUser(c *gin.Context) {
	uid := c.Param("uid")
	user := new(model.FaceUser)
	if err := config.DB.Where("`Uid`=?", uid).First(user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"err": "未找到该uuid用户"})
		return
	}

	c.JSON(http.StatusOK, user)
	return
}

//  godoc
// @Summary Create or Update user
// @Description create user if uid not exists else update
// @Accept  multipart/form-data
// @Produce  json
// @param uid formData string true "user id"
// @param name formData string false "name"
// @Param faceFile formData file false "人脸照片"
// @Param idFile formData file false "身份证人面照"
// @param gid formData string false "group id"
// @param faceFeature formData string false "人脸特征文件, binary格式" format(binary)
// @param faceImagePath formData string false "人脸照片路径（服务器已存在的相对路径）"
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
		group := model.FaceGroup{}
		config.DB.Where("`Gid`=?", U.Gid).First(&group)
		user.Groups = append(user.Groups, group)
	}

	var wg sync.WaitGroup

	if U.IdFile != nil {
		idFile, _ := U.IdFile.Open()
		defer idFile.Close()
		md5 := utils.MD5sum(idFile)
		user.IdImagePath = fmt.Sprintf("%s/%s/%s/%s.jpg", md5[0:2], md5[2:4], md5[4:6], md5)
		wg.Add(1)
		go func() {
			if err := utils.SaveFile(idFile, path.Join(config.RegDir, user.IdImagePath)); err != nil {
				config.Logger.Error("保存身份证人脸照片失败", zap.String("文件", user.IdImagePath), zap.Error(err))
			}
			wg.Done()
		}()
	}

	if U.FaceFile != nil {
		faceFile, err := U.FaceFile.Open()
		if err != nil {
			config.Logger.Error("打开文件失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"err": err.Error()})
			return
		}

		defer faceFile.Close()

		// 计算文件的md5
		md5 := utils.MD5sum(faceFile)
		user.FaceImagePath = fmt.Sprintf("%s/%s/%s/%s.jpg", md5[0:2], md5[2:4], md5[4:6], md5)
		user.FaceFeature, err = recognition.FeatureByteArr(faceFile)
		if err != nil {
			config.Logger.Warn("图片提取人脸特征失败", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			return
		} else {
			faceFile.Seek(0, io.SeekStart) // seek到0
			wg.Add(1)
			go func() {
				if err := utils.SaveFile(faceFile, path.Join(config.RegDir, user.FaceImagePath)); err != nil {
					config.Logger.Error("保存人脸照片失败", zap.String("文件", user.FaceImagePath), zap.Error(err))
				}
				wg.Done()
			}()
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
			buf := new(bytes.Buffer)
			io.Copy(buf, ff)
			user.FaceFeature = buf.Bytes()
		} else {
			if len(user.FaceImagePath) > 0 {
				var err error
				user.FaceFeature, err = recognition.FeatureByteArr(config.RegDir + "/" + user.FaceImagePath)
				if err != nil {
					config.Logger.Warn("图片提取人脸特征失败", zap.Error(err))
					c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
					return
				}
			}
		}
	}

	if user.FaceFeature != nil {
		lock := cache.Mutex(user, config.MultiPoint)
		lock.Lock()
		if err := config.DB.Where("Uid=?", U.Uid).Assign(user).FirstOrCreate(&user).Error; err != nil {
			config.Logger.Error("保存数据失败", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			lock.Unlock()
			return
		}
		cache.UpdateUserFeature(user.Uid, user.FaceFeature)
		lock.Unlock()
	} else {
		if err := config.DB.Where("Uid=?", U.Uid).Assign(user).FirstOrCreate(&user).Error; err != nil {
			config.Logger.Error("保存数据失败", zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
			return
		}
	}
	wg.Wait()

	for _, g := range U.Gid {
		cache.AddGroupFeatures(g, user.Uid)
	}

	c.JSON(http.StatusCreated, model.FaceUser{
		Uid:           user.Uid,
		Name:          user.Name,
		FaceImagePath: user.FaceImagePath,
		IdImagePath:   user.IdImagePath,
	})
}

//  godoc
// @Summary Match a user
// @Description post a faceFile to match a user in a group and save the faceFile.
// @Accept  multipart/form-data
// @Param faceFile formData file true "faceFile"
// @param gid formData string true "group id"
// @Produce  json
// @Success 201 {object} response.FaceMatchResult
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

	// matches, err := eng.SearchN(ff, group.FaceFeatures(), 1, 0.8)

	matches, err := eng.SearchN(ff, cache.GetGroupFeatures(group), 1, 0.8)
	if err != nil {
		config.Logger.Error("人脸查找发生错误", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"err": "人脸匹配发生错误"})
		return
	}
	if matches == nil || len(matches) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"err": "未匹配到任何人脸"})
		return
	}
	// 异步存储人脸
	md5 := utils.MD5sum(ff)
	vfp := fmt.Sprintf("%s/%s/%s/%s.jpg", md5[0:2], md5[2:4], md5[4:6], md5)
	go func(f string) {
		dst := path.Join(config.VerDir, f)
		utils.CreateDir(path.Dir(dst))
		c.SaveUploadedFile(M.FaceFile, dst)
	}(vfp)

	result := response.FaceMatchResult{
		matches[0], vfp,
	}
	config.Logger.Info("匹配成功", zap.Any("结果", result))
	c.JSON(http.StatusOK, result)
}
