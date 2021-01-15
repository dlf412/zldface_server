// 维护分组信息
package v1

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"zldface_server/cache"
	"zldface_server/config"
	"zldface_server/model"
	"zldface_server/model/request"
)

// CreateGroup godoc
// @Summary Greate a group
// @Description create a group using gid and name
// @Accept  json
// @Produce  json
// @Param data body request.FaceGroup true "group id, group name"
// @Success 201 {object} model.FaceGroup
// @Router /groups/v1 [post]
func CreateGroup(c *gin.Context) {

	var G request.FaceGroup

	if err := c.ShouldBindJSON(&G); err != nil {
		config.Logger.Info(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	group := &model.FaceGroup{Gid: G.Gid, Name: G.Name}
	if err := config.DB.Create(&group).Error; err != nil {
		config.Logger.Info(err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"err": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, group)
	config.Logger.Info("create group ok", zap.Any("group", group))
}

//  godoc
// @Summary Create group users
// @Description add users to a group
// @Accept  json
// @Produce  json
// @Param data body request.FaceGroupUser true "分组id, 用户uid列表"
// @Success 201 {string} string "{"msg":"添加成功"}"
// @Router /group/users/v1 [post]
func CreateGroupUser(c *gin.Context) {

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
		c.JSON(http.StatusBadRequest, gin.H{"err": "group不存在"})
		return
	}
	if len(G.Uids) > 0 {
		config.DB.Where("`Uid` IN ?", G.Uids).Find(&users)
		lock := cache.Mutex(group, config.MultiPoint)
		lock.Lock()
		defer lock.Unlock()
		err := ass.Append(users)
		if err != nil {
			config.Logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"err": "数据库操作错误"})
			return
		}
		cache.AddGroupFeatures(G.Gid, G.Uids...)
	}
	c.JSON(http.StatusCreated, gin.H{"msg": "添加成功"})
	config.Logger.Info("create group users ok", zap.Any("group", G))
}

//  godoc
// @Summary Delete group users
// @Description delete users from a group
// @Accept  json
// @Produce  json
// @Param data body request.FaceGroupUser true "分组id, 用户uid列表"
// @Success 200 {string} string "{"msg":"删除成功"}"
// @Router /group/users/v1 [delete]
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
		lock := cache.Mutex(group, config.MultiPoint)
		lock.Lock()
		defer lock.Unlock()
		err := config.DB.Model(&group).Association("Users").Delete(users)
		if err != nil {
			config.Logger.Error(err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{"err": "数据库操作错误"})
			return
		}
		cache.DelGroupFeatures(G.Gid, G.Uids...)
	}
	c.JSON(http.StatusOK, gin.H{"msg": "删除成功"})
	config.Logger.Info("delete group users ok", zap.Any("group", G))
}
