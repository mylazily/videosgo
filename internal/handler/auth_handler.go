package handler

import (
	"net/http"

	"videosgo/internal/service"

	"github.com/gin-gonic/gin"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req service.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "注册成功",
		"data":    user.ToResponse(),
	})
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    400,
			"message": err.Error(),
		})
		return
	}

	tokens, user, err := h.authService.Login(&req)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"code":    401,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "登录成功",
		"data": gin.H{
			"token":         tokens.AccessToken,
			"refresh_token": tokens.RefreshToken,
			"expires_in":    tokens.ExpiresIn,
			"token_type":    tokens.TokenType,
			"user":          user.ToResponse(),
		},
	})
}

// RefreshToken 刷新 Token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// TODO: 实现刷新逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "刷新成功",
	})
}
