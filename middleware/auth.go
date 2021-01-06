package middleware

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"go.uber.org/zap"
	"net/http"
	"strings"
	"zldface_server/config"
)

var ctx = context.Background()

// 智链达Sid认证  Authorization:SID $sid
func ZldAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
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
		user, err := config.RedisCli.Get(ctx, token).Result()
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
			config.Logger.Info(fmt.Sprintf("auth认证成功"), zap.String("sid", token), zap.String("user_info", user))
			c.Next()
		}
	}
}
