// 允许最大的并发数, 适用于某些耗时的接口
package middleware

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

// 最大并发数
func MaxAllowed(n int) gin.HandlerFunc {
	sem := make(chan struct{}, n)
	acquire := func() bool {
		select {
		case <-time.After(time.Millisecond * 500):
			return false
		case sem <- struct{}{}:
			return true
		}
	}
	release := func() { <-sem }
	return func(c *gin.Context) {
		if acquire() {
			defer release()
			c.Next()
		} else {
			c.JSON(http.StatusServiceUnavailable, gin.H{"detail": "服务器繁忙，请稍候再试"})
			c.Abort()
			return
		}
	}
}
