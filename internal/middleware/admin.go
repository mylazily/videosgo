package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"videosgo/pkg/response"
)

// RequireAdmin 管理员权限中间件
// 必须在 Auth 中间件之后使用，因为依赖 c.Get("is_admin")
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		isAdmin, exists := c.Get("is_admin")
		if !exists {
			response.Unauthorized(c, "未登录")
			c.Abort()
			return
		}

		isAdminBool, ok := isAdmin.(bool)
		if !ok {
			response.Unauthorized(c, "用户权限信息无效")
			c.Abort()
			return
		}

		if !isAdminBool {
			response.Error(c, http.StatusForbidden, "需要管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}
