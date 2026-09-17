package utils

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// 统一响应格式：{"code": 0, "message": "ok", "data": ...}
// code = 0 表示成功；非 0 为业务错误码（见 errors.go）。

// OK 成功响应（200）。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"code": 0, "message": "ok", "data": data})
}

// Created 创建成功响应（201）。
func Created(c *gin.Context, message string, data any) {
	c.JSON(http.StatusCreated, gin.H{"code": 0, "message": message, "data": data})
}

// Fail 错误响应。
func Fail(c *gin.Context, appErr *AppError) {
	c.JSON(appErr.HTTPStatus, gin.H{"code": appErr.Code, "message": appErr.Message, "data": nil})
}

// HandleError 统一错误出口：记录日志（4xx 为 warn，5xx 为 error）并返回标准化响应。
func HandleError(c *gin.Context, err error) {
	appErr := AsAppError(err)
	if appErr == ErrInternal {
		slog.Error("服务器内部错误", "path", c.FullPath(), "error", err)
	} else {
		slog.Warn("请求未通过", "method", c.Request.Method, "path", c.FullPath(),
			"code", appErr.Code, "message", appErr.Message)
	}
	Fail(c, appErr)
}
