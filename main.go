package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"time"
	"zldface_server/cache"
	"zldface_server/config"
	"zldface_server/router"
)

type server interface {
	ListenAndServe() error
}

func initServer(address string, router *gin.Engine) server {
	return &http.Server{
		Addr:           address,
		Handler:        router,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

func runserver() {
	Router := router.Routers()
	//request.RegisterValidations()
	//Router.Static("/form-generator", "./resource/page")
	address := fmt.Sprintf(":%d", config.Config.System.Addr)
	s := initServer(address, Router)
	time.Sleep(10 * time.Microsecond)
	config.Logger.Info("server run success on ", zap.String("address", address))
	config.Logger.Error(s.ListenAndServe().Error())
}

// @title 智链达人脸录入和识别服务API
// @version 1.0
// @description This a face recognition server using arcsoft face engine
// @contact.name DengLingfei
// @contact.email denglingfei@zlddata.cn
// @license.name Apache2.0
// @host localhost:8888
// @BasePath /face/
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {
	config.Logger.Info(fmt.Sprintf("%v", config.Config))
	cache.BeRun()
	runserver()

	db, _ := config.DB.DB()
	defer db.Close()
	if config.RedisCli != nil {
		defer config.RedisCli.Close()
	}
}
