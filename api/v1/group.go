// 维护分组信息
package v1

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"zldface_server/config"
	"zldface_server/model"
	"zldface_server/model/request"
)

func CreateGroup(c *gin.Context) {

	var G request.FaceGroup

	if err := c.ShouldBindJSON(&G); err != nil {
		config.Logger.Info(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// group := &model.FaceGroup{Gid: G.Gid, Name:G.Name}
	if err := config.DB.Create(&G).Error; err != nil {
		config.Logger.Info(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"message": "创建成功"})
	config.Logger.Info("create group ok", zap.Any("group", G))
}

func GreateGroupUser(c *gin.Context) {

	var G request.FaceGroupUser

	if err := c.ShouldBindJSON(&G); err != nil {
		config.Logger.Info(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	config.Logger.Info("收到请求", zap.Any("group:users", G))

	// db.Exec("INSERT INTO `face_group_users` (`face_group_id`,`face_user_id`) VALUES (?,?) (?,?)", )
	group := model.FaceGroup{}
	users := []model.FaceUser{}
	ass := config.DB.Where("Gid = ?", G.Gid).First(&group).Association("Users")
	if group.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group不存在"})
		return
	}
	if len(G.Uids) > 0 {
		config.DB.Where("`Uid` IN ?", G.Uids).Find(&users)
		ass.Append(users)
	}
	c.JSON(http.StatusCreated, gin.H{"message": "添加成功"})
	config.Logger.Info("create group users ok", zap.Any("group", G))
}

func DeleteGroupUser(c *gin.Context) {

	var G request.FaceGroupUser
	if err := c.ShouldBindJSON(&G); err != nil {
		config.Logger.Info(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	config.Logger.Info("收到请求", zap.Any("group:users", G))

	// db.Exec("INSERT INTO `face_group_users` (`face_group_id`,`face_user_id`) VALUES (?,?) (?,?)", )
	group := model.FaceGroup{}
	users := []model.FaceUser{}

	config.DB.Where("Gid = ?", G.Gid).First(&group)
	if group.ID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group不存在"})
		return
	}
	if len(G.Uids) > 0 {
		config.DB.Where("`Uid` IN ?", G.Uids).Find(&users)
		config.DB.Model(&group).Association("Users").Delete(users)
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
	config.Logger.Info("delete group users ok", zap.Any("group", G))
}
