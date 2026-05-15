// Package middleware HTTP 中间件
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	jwtpkg "github.com/mylazily/videosgo/pkg/jwt"
	"github.com/mylazily/videosgo/pkg/response"
)

// Auth JWT 认证中间件
func Auth(jwtMgr *jwtpkg.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 Header 获取令牌
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "请先登录")
			c.Abort()
			return
		}

		// 支持 Bearer token 格式
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "令牌格式错误")
			c.Abort()
			return
		}

		tokenString := parts[1]
		claims, err := jwtMgr.ParseToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "令牌无效或已过期")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("is_admin", claims.IsAdmin)

		c.Next()
	}
}

// AdminRequired 管理员权限中间件（需配合 Auth 使用）
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists || !isAdmin.(bool) {
			response.Forbidden(c, "需要管理员权限")
			c.Abort()
			return
		}
		c.Next()
	}
}
