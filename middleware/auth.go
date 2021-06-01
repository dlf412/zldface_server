package middleware

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"zldface_server/config"
)

// Sid认证  Authorization:SID $sid
func ZldAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 简单的认为 host为 localhost 则不需要认证，内部调用
		if c.Request.Host == fmt.Sprintf("localhost:%d", config.Config.System.Addr) {
			c.Next()
			return
		}
		if config.RedisCli == nil { // 单点模式ZldAuth无效
			c.Next()
			return
		}
		sid := c.Request.Header.Get("Authorization")
		if len(sid) == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"detail": "未传SID"})
			c.Abort()
			return
		}
		s := strings.Split(sid, " ")
		if len(s) < 2 || s[0] != "SID" {
			c.JSON(http.StatusForbidden, gin.H{"detail": "非法SID"})
			c.Abort()
			return
		}
		token := ":1:" + s[1]
		// 从redis cache里查找对应的token, 看是否存在
		_, err := config.RedisCli.Get(config.Rctx, token).Result()
		if err == redis.Nil {
			c.JSON(http.StatusForbidden, gin.H{"detail": "无效SID"})
			c.Abort()
			return
		} else if err != nil {
			config.Logger.Error("服务器发生错误", zap.Error(err))
			c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "服务器繁忙，请稍候再试"})
			c.Abort()
			return
		} else {
			config.Logger.Info(fmt.Sprintf("auth认证成功"), zap.String("sid", token))
			c.Next()
		}
	}
}
