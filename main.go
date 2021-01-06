package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"net/http"
	"time"
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
	//Router.Static("/form-generator", "./resource/page")
	address := fmt.Sprintf(":%d", config.Config.System.Addr)
	s := initServer(address, Router)
	time.Sleep(10 * time.Microsecond)
	config.Logger.Info("server run success on ", zap.String("address", address))
	config.Logger.Error(s.ListenAndServe().Error())
}

func main() {
	config.Logger.Info(fmt.Sprintf("%v", config.Config))
	config.Logger.Info(config.RedisCli.String())
	config.Logger.Info(fmt.Sprintf("%v", *config.DB))
	config.Logger.Info(config.RegDir)
	config.Logger.Info(config.VerDir)

	runserver()

	db, _ := config.DB.DB()
	defer db.Close()
	defer config.RedisCli.Close()
}

//import "github.com/gin-gonic/gin"
//
//func main() {
//	r := gin.Default()
//	r.GET("/ping", func(c *gin.Context) {
//		c.JSON(200, gin.H{
//			"message": "pong",
//		})
//	})
//	r.Run() // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
//}
