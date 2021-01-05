package main

import (
	"zldface_server/config"
)

func main() {
	config.Logger.Info(config.RedisCli.String())
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