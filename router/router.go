package router

import (
	"github.com/gin-gonic/gin"
	"zldface_server/config"
	"zldface_server/middleware"
)

func Routers() *gin.Engine {
	var Router = gin.Default()
	//Router.StaticFS(global.GVA_CONFIG.Local.Path, http.Dir(global.GVA_CONFIG.Local.Path))
	// Router.Use(middleware.LoadTls())  // 打开就能玩https了
	config.Logger.Info("use middleware logger")
	// 跨域
	Router.Use(middleware.Cors())
	config.Logger.Info("use middleware cors")
	//Router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	//global.GVA_LOG.Info("register swagger handler")
	// 方便统一添加路由组前缀 多服务器上线使用
	//PublicGroup := Router.Group("") // 不需要auth
	//{
	//	router.InitBaseRouter(PublicGroup) // 注册基础功能路由 不做鉴权
	//}
	PrivateGroup := Router.Group("")
	if !config.Debug {
		if config.Config.Auth == "ZldAuth" {
			PrivateGroup.Use(middleware.ZldAuth())
		}
	}
	{
		InitFaceRouter(PrivateGroup) // 注册功能api路由
	}
	config.Logger.Info("router register success")
	return Router
}
