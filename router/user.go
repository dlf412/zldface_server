package router

import (
	"github.com/gin-gonic/gin"
	v1 "zldface_server/api/v1"
)

func InitUserRouter(Router *gin.RouterGroup) {
	UserRouter := Router.Group("users")
	{
		UserRouter.POST("v1", v1.CreateUser) // 创建Api
		//FaceRouter.DELETE("v1", v1.DeleteFace)   // 删除Api
		//FaceRouter.GET("v1", v1.GetFaceList) // 获取Api列表
		//FaceRouter.GET("v1/:id", v1.GetFaceById) // 获取单条Api消息
		//FaceRouter.PATCH("v1/:id", v1.UpdateFace)   // 更新api
	}

	UserMatchRouter := Router.Group("user/match")
	{
		UserMatchRouter.POST("v1", v1.MatchUser)
	}
}
