package middleware

import (
	"fmt"
	"strings"

	"videosgo/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// Auth JWT 认证中间件
func Auth(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "Authorization header required")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.Unauthorized(c, "Invalid authorization format")
			c.Abort()
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})

		if err != nil || !token.Valid {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			// 类型安全的 user_id 提取
			var userID string
			switch v := claims["user_id"].(type) {
			case string:
				userID = v
			case float64:
				// JWT 数字类型可能是 float64
				userID = fmt.Sprintf("%.0f", v)
			default:
				if v != nil {
					userID = fmt.Sprintf("%v", v)
				}
			}
			c.Set("user_id", userID)

			// 类型安全的 email 提取
			var email string
			if v, ok := claims["email"].(string); ok {
				email = v
			}
			c.Set("email", email)
		}

		c.Next()
	}
}
