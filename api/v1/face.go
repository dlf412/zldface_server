package v1

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"zldface_server/config"
)

func CreateFace(c *gin.Context) {
	c.JSON(http.StatusCreated, gin.H{"hello": "world"})
	config.Logger.Info("create face ok")
}
