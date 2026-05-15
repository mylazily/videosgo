package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// 敏感参数列表（不记录到日志中）
var sensitiveParams = map[string]bool{
	"password":     true,
	"old_password": true,
	"new_password": true,
	"token":        true,
	"secret":       true,
	"api_key":      true,
}

// Logging 请求日志中间件（过滤敏感参数）
func Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 处理请求
		c.Next()

		// 计算耗时
		latency := time.Since(start)
		status := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method

		// 过滤查询参数中的敏感信息
		filteredQuery := filterSensitiveParams(query)

		log.Printf("[请求] %s | %3d | %13v | %s | %s %s",
			clientIP,
			status,
			latency,
			method,
			path,
			filteredQuery,
		)
	}
}

// filterSensitiveParams 过滤查询参数中的敏感信息
func filterSensitiveParams(query string) string {
	if query == "" {
		return ""
	}

	// 简单实现：隐藏敏感参数的值
	parts := splitQuery(query)
	for i, part := range parts {
		for param := range sensitiveParams {
			if len(part) > len(param)+1 && part[:len(param)+1] == param+"=" {
				parts[i] = param + "=***"
				break
			}
		}
	}

	return joinQuery(parts)
}

// splitQuery 分割查询字符串
func splitQuery(query string) []string {
	var parts []string
	start := 0
	for i := 0; i < len(query); i++ {
		if query[i] == '&' {
			parts = append(parts, query[start:i])
			start = i + 1
		}
	}
	if start < len(query) {
		parts = append(parts, query[start:])
	}
	return parts
}

// joinQuery 连接查询字符串
func joinQuery(parts []string) string {
	result := ""
	for i, part := range parts {
		if i > 0 {
			result += "&"
		}
		result += part
	}
	return result
}
