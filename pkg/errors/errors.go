// Package errors 统一错误定义和处理
package errors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ErrorCode 错误码类型
type ErrorCode int

// 错误码定义
const (
	// 通用错误 (1000-1999)
	ErrCodeUnknown        ErrorCode = 1000 // 未知错误
	ErrCodeInternal       ErrorCode = 1001 // 内部服务器错误
	ErrCodeInvalidParams  ErrorCode = 1002 // 参数错误
	ErrCodeUnauthorized   ErrorCode = 1003 // 未授权
	ErrCodeForbidden      ErrorCode = 1004 // 禁止访问
	ErrCodeNotFound       ErrorCode = 1005 // 资源不存在
	ErrCodeTooManyRequest ErrorCode = 1006 // 请求过于频繁
	ErrCodeTimeout        ErrorCode = 1007 // 请求超时

	// 用户相关错误 (2000-2999)
	ErrCodeUserNotFound     ErrorCode = 2000 // 用户不存在
	ErrCodeUserExists       ErrorCode = 2001 // 用户已存在
	ErrCodeInvalidPassword  ErrorCode = 2002 // 密码错误
	ErrCodeInvalidToken     ErrorCode = 2003 // Token 无效
	ErrCodeTokenExpired     ErrorCode = 2004 // Token 已过期
	ErrCodeInvalidUsername  ErrorCode = 2005 // 用户名格式错误
	ErrCodeInvalidEmail     ErrorCode = 2006 // 邮箱格式错误
	ErrCodeWeakPassword     ErrorCode = 2007 // 密码强度不足

	// 视频相关错误 (3000-3999)
	ErrCodeVideoNotFound    ErrorCode = 3000 // 视频不存在
	ErrCodeVideoExists      ErrorCode = 3001 // 视频已存在
	ErrCodeVideoUnavailable ErrorCode = 3002 // 视频不可用
	ErrCodeEpisodeNotFound  ErrorCode = 3003 // 剧集不存在

	// 采集相关错误 (4000-4999)
	ErrCodeCollectFailed    ErrorCode = 4000 // 采集失败
	ErrCodeSourceNotFound   ErrorCode = 4001 // 采集源不存在
	ErrCodeSourceDisabled   ErrorCode = 4002 // 采集源已禁用
	ErrCodeAPIError         ErrorCode = 4003 // API 错误
	ErrCodeParseError       ErrorCode = 4004 // 解析错误
	ErrCodeProbeFailed      ErrorCode = 4005 // 探针失败

	// 数据库相关错误 (5000-5999)
	ErrCodeDBError          ErrorCode = 5000 // 数据库错误
	ErrCodeDBConnectFailed  ErrorCode = 5001 // 数据库连接失败
	ErrCodeDBQueryFailed    ErrorCode = 5002 // 数据库查询失败

	// Redis 相关错误 (6000-6999)
	ErrCodeRedisError       ErrorCode = 6000 // Redis 错误
	ErrCodeCacheError       ErrorCode = 6001 // 缓存错误
)

// ErrorCodeMapping 错误码与 HTTP 状态码映射
var ErrorCodeMapping = map[ErrorCode]int{
	ErrCodeUnknown:        http.StatusInternalServerError,
	ErrCodeInternal:       http.StatusInternalServerError,
	ErrCodeInvalidParams:  http.StatusBadRequest,
	ErrCodeUnauthorized:   http.StatusUnauthorized,
	ErrCodeForbidden:      http.StatusForbidden,
	ErrCodeNotFound:       http.StatusNotFound,
	ErrCodeTooManyRequest: http.StatusTooManyRequests,
	ErrCodeTimeout:        http.StatusRequestTimeout,

	ErrCodeUserNotFound:     http.StatusNotFound,
	ErrCodeUserExists:       http.StatusConflict,
	ErrCodeInvalidPassword:  http.StatusUnauthorized,
	ErrCodeInvalidToken:     http.StatusUnauthorized,
	ErrCodeTokenExpired:     http.StatusUnauthorized,
	ErrCodeInvalidUsername:  http.StatusBadRequest,
	ErrCodeInvalidEmail:     http.StatusBadRequest,
	ErrCodeWeakPassword:     http.StatusBadRequest,

	ErrCodeVideoNotFound:    http.StatusNotFound,
	ErrCodeVideoExists:      http.StatusConflict,
	ErrCodeVideoUnavailable: http.StatusServiceUnavailable,
	ErrCodeEpisodeNotFound:  http.StatusNotFound,

	ErrCodeCollectFailed:    http.StatusInternalServerError,
	ErrCodeSourceNotFound:   http.StatusNotFound,
	ErrCodeSourceDisabled:   http.StatusForbidden,
	ErrCodeAPIError:         http.StatusBadGateway,
	ErrCodeParseError:       http.StatusUnprocessableEntity,
	ErrCodeProbeFailed:      http.StatusServiceUnavailable,

	ErrCodeDBError:         http.StatusInternalServerError,
	ErrCodeDBConnectFailed: http.StatusServiceUnavailable,
	ErrCodeDBQueryFailed:   http.StatusInternalServerError,

	ErrCodeRedisError: http.StatusInternalServerError,
	ErrCodeCacheError: http.StatusInternalServerError,
}

// ErrorCodeMessage 错误码与错误消息映射
var ErrorCodeMessage = map[ErrorCode]string{
	ErrCodeUnknown:        "未知错误",
	ErrCodeInternal:       "服务器内部错误",
	ErrCodeInvalidParams:  "请求参数错误",
	ErrCodeUnauthorized:   "未授权访问",
	ErrCodeForbidden:      "禁止访问",
	ErrCodeNotFound:       "资源不存在",
	ErrCodeTooManyRequest: "请求过于频繁，请稍后再试",
	ErrCodeTimeout:        "请求超时",

	ErrCodeUserNotFound:     "用户不存在",
	ErrCodeUserExists:       "用户已存在",
	ErrCodeInvalidPassword:  "密码错误",
	ErrCodeInvalidToken:     "登录凭证无效",
	ErrCodeTokenExpired:     "登录已过期，请重新登录",
	ErrCodeInvalidUsername:  "用户名格式错误",
	ErrCodeInvalidEmail:     "邮箱格式错误",
	ErrCodeWeakPassword:     "密码强度不足",

	ErrCodeVideoNotFound:    "视频不存在",
	ErrCodeVideoExists:      "视频已存在",
	ErrCodeVideoUnavailable: "视频暂不可用",
	ErrCodeEpisodeNotFound:  "剧集不存在",

	ErrCodeCollectFailed:    "采集任务执行失败",
	ErrCodeSourceNotFound:   "采集源不存在",
	ErrCodeSourceDisabled:   "采集源已禁用",
	ErrCodeAPIError:         "外部 API 调用失败",
	ErrCodeParseError:       "数据解析失败",
	ErrCodeProbeFailed:      "链接探针检测失败",

	ErrCodeDBError:         "数据库操作失败",
	ErrCodeDBConnectFailed: "数据库连接失败",
	ErrCodeDBQueryFailed:   "数据库查询失败",

	ErrCodeRedisError: "缓存服务异常",
	ErrCodeCacheError: "缓存操作失败",
}

// AppError 应用错误结构
type AppError struct {
	Code    ErrorCode `json:"code"`    // 错误码
	Message string    `json:"message"` // 错误消息
	Detail  string    `json:"detail"`  // 详细错误信息（仅开发环境返回）
	Err     error     `json:"-"`       // 原始错误
}

// Error 实现 error 接口
func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("[%d] %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("[%d] %s", e.Code, e.Message)
}

// Unwrap 返回原始错误
func (e *AppError) Unwrap() error {
	return e.Err
}

// WithDetail 添加详细错误信息
func (e *AppError) WithDetail(detail string) *AppError {
	e.Detail = detail
	return e
}

// WithError 添加原始错误
func (e *AppError) WithError(err error) *AppError {
	e.Err = err
	return e
}

// HTTPStatus 返回对应的 HTTP 状态码
func (e *AppError) HTTPStatus() int {
	if status, ok := ErrorCodeMapping[e.Code]; ok {
		return status
	}
	return http.StatusInternalServerError
}

// JSON 返回 JSON 格式的错误信息
func (e *AppError) JSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

// New 创建新的应用错误
func New(code ErrorCode, message ...string) *AppError {
	msg := ErrorCodeMessage[code]
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return &AppError{
		Code:    code,
		Message: msg,
	}
}

// NewWithError 创建包含原始错误的应用错误
func NewWithError(code ErrorCode, err error, message ...string) *AppError {
	return New(code, message...).WithError(err)
}

// NewWithDetail 创建包含详细信息的应用错误
func NewWithDetail(code ErrorCode, detail string, message ...string) *AppError {
	return New(code, message...).WithDetail(detail)
}

// Is 判断错误是否匹配
func Is(err error, code ErrorCode) bool {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

// GetCode 从错误中获取错误码
func GetCode(err error) ErrorCode {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return ErrCodeUnknown
}

// GetHTTPStatus 从错误中获取 HTTP 状态码
func GetHTTPStatus(err error) int {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr.HTTPStatus()
	}
	return http.StatusInternalServerError
}

// Wrap 包装错误
func Wrap(err error, code ErrorCode, message ...string) error {
	if err == nil {
		return nil
	}
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return NewWithError(code, err, message...)
}

// 快捷创建函数

// Internal 创建内部错误
func Internal(err error, message ...string) *AppError {
	return NewWithError(ErrCodeInternal, err, message...)
}

// InvalidParams 创建参数错误
func InvalidParams(detail string, message ...string) *AppError {
	return NewWithDetail(ErrCodeInvalidParams, detail, message...)
}

// Unauthorized 创建未授权错误
func Unauthorized(message ...string) *AppError {
	return New(ErrCodeUnauthorized, message...)
}

// Forbidden 创建禁止访问错误
func Forbidden(message ...string) *AppError {
	return New(ErrCodeForbidden, message...)
}

// NotFound 创建资源不存在错误
func NotFound(resource string, message ...string) *AppError {
	msg := fmt.Sprintf("%s不存在", resource)
	if len(message) > 0 && message[0] != "" {
		msg = message[0]
	}
	return New(ErrCodeNotFound).WithDetail(msg)
}

// DBError 创建数据库错误
func DBError(err error, message ...string) *AppError {
	return NewWithError(ErrCodeDBError, err, message...)
}

// APIError 创建 API 错误
func APIError(err error, message ...string) *AppError {
	return NewWithError(ErrCodeAPIError, err, message...)
}

// CollectError 创建采集错误
func CollectError(err error, message ...string) *AppError {
	return NewWithError(ErrCodeCollectFailed, err, message...)
}
