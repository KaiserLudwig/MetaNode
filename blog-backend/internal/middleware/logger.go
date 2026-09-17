package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger 请求日志中间件：记录方法、路径、状态码、客户端 IP 与耗时。
func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		slog.Info("HTTP 请求",
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"ip", c.ClientIP(),
			"cost", time.Since(start).Round(time.Microsecond).String(),
		)
	}
}

// Recovery panic 恢复中间件：把未捕获的 panic 转成标准 500 响应，避免进程崩溃。
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, err any) {
		slog.Error("panic 已恢复", "path", c.FullPath(), "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    50000,
			"message": "服务器内部错误",
			"data":    nil,
		})
		c.Abort()
	})
}
