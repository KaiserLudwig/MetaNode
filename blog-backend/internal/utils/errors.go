// Package utils 通用工具：统一响应、业务错误、JWT。
package utils

import (
	"errors"

	"gorm.io/gorm"
)

// AppError 业务错误：包含对外错误码、HTTP 状态码与用户可读信息。
type AppError struct {
	Code       int    // 业务错误码（0 为成功）
	HTTPStatus int    // HTTP 状态码
	Message    string // 用户可读信息
}

func (e *AppError) Error() string { return e.Message }

// 常用业务错误。
var (
	ErrBadRequest   = &AppError{Code: 40000, HTTPStatus: 400, Message: "请求参数错误"}
	ErrUnauthorized = &AppError{Code: 40100, HTTPStatus: 401, Message: "未登录或登录已过期"}
	ErrForbidden    = &AppError{Code: 40300, HTTPStatus: 403, Message: "无权执行该操作"}
	ErrNotFound     = &AppError{Code: 40400, HTTPStatus: 404, Message: "资源不存在"}
	ErrConflict     = &AppError{Code: 40900, HTTPStatus: 409, Message: "资源冲突"}
	ErrInternal     = &AppError{Code: 50000, HTTPStatus: 500, Message: "服务器内部错误"}
)

// 带自定义信息的错误构造器。
func BadRequest(msg string) *AppError   { return &AppError{Code: 40000, HTTPStatus: 400, Message: msg} }
func Unauthorized(msg string) *AppError { return &AppError{Code: 40100, HTTPStatus: 401, Message: msg} }
func Forbidden(msg string) *AppError    { return &AppError{Code: 40300, HTTPStatus: 403, Message: msg} }
func NotFound(msg string) *AppError     { return &AppError{Code: 40400, HTTPStatus: 404, Message: msg} }
func Conflict(msg string) *AppError     { return &AppError{Code: 40900, HTTPStatus: 409, Message: msg} }

// AsAppError 把任意错误归一化为 AppError：
// 业务错误原样返回，GORM 未找到记录映射为 404，其余视为 500。
func AsAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return ErrInternal
}
