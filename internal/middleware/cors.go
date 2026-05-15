package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CORS 跨域中间件（支持通配符域名）
func CORS(allowedOrigins []string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader("Origin")

		// 检查是否允许该来源
		allowed := false
		for _, ao := range allowedOrigins {
			if ao == "*" || ao == origin {
				allowed = true
				break
			}
			// 支持通配符后缀匹配，如 *.example.com
			if strings.HasPrefix(ao, "*.") && strings.HasSuffix(origin, ao[1:]) {
				allowed = true
				break
			}
		}

		if allowed {
			if len(allowedOrigins) == 1 && allowedOrigins[0] == "*" {
				c.Header("Access-Control-Allow-Origin", "*")
			} else {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Vary", "Origin")
			}
			c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS, PATCH")
			c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Request-ID")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
