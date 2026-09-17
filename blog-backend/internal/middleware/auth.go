// Package middleware Gin 中间件：JWT 认证、请求日志、panic 恢复。
package middleware

import (
	"strings"

	"blog-backend/internal/utils"

	"github.com/gin-gonic/gin"
)

// 上下文键。
const (
	ctxUserIDKey   = "authUserID"
	ctxUsernameKey = "authUsername"
)

// Auth JWT 认证中间件：校验 Authorization: Bearer <token>，
// 通过后把用户 ID 与用户名写入上下文。
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			utils.Fail(c, utils.Unauthorized("缺少认证令牌，请在请求头携带 Authorization: Bearer <token>"))
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := utils.ParseToken(secret, tokenStr)
		if err != nil {
			utils.Fail(c, utils.Unauthorized("认证令牌无效或已过期"))
			c.Abort()
			return
		}
		c.Set(ctxUserIDKey, claims.UserID)
		c.Set(ctxUsernameKey, claims.Username)
		c.Next()
	}
}

// UserID 从上下文取出当前登录用户 ID。
func UserID(c *gin.Context) uint {
	v, _ := c.Get(ctxUserIDKey)
	id, _ := v.(uint)
	return id
}

// Username 从上下文取出当前登录用户名。
func Username(c *gin.Context) string {
	v, _ := c.Get(ctxUsernameKey)
	s, _ := v.(string)
	return s
}
