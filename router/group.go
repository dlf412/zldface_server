package router

import (
	"github.com/gin-gonic/gin"
	v1 "zldface_server/api/v1"
)

func InitGroupRouter(Router *gin.RouterGroup) {
	GroupRouter := Router.Group("groups")
	{
		GroupRouter.POST("v1", v1.CreateGroup) // 创建Api
		//FaceRouter.DELETE("v1", v1.DeleteFace)   // 删除Api
		//FaceRouter.GET("v1", v1.GetFaceList) // 获取Api列表
		//FaceRouter.GET("v1/:id", v1.GetFaceById) // 获取单条Api消息
		//FaceRouter.PATCH("v1/:id", v1.UpdateFace)   // 更新api
	}
	GroupUserRouter := Router.Group("group/users")
	{
		GroupUserRouter.POST("v1", v1.CreateGroupUser)
		GroupUserRouter.DELETE("v1", v1.DeleteGroupUser)
	}
}
