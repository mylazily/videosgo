package handler

import (
	"errors"
	"net/http"

	"videosgo/internal/service"
	"videosgo/pkg/response"

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
		response.BadRequest(c, err.Error())
		return
	}

	user, err := h.authService.Register(&req)
	if err != nil {
		if errors.Is(err, service.ErrUserAlreadyExists) || errors.Is(err, service.ErrInvalidInput) {
			response.BadRequest(c, err.Error())
			return
		}
		response.InternalError(c, "注册失败，请稍后重试")
		return
	}

	response.SuccessWithMessage(c, "注册成功", user.ToResponse())
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req service.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	tokens, user, err := h.authService.Login(&req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) || errors.Is(err, service.ErrInvalidInput) {
			response.Unauthorized(c, err.Error())
			return
		}
		response.InternalError(c, "登录失败，请稍后重试")
		return
	}

	response.Success(c, gin.H{
		"token":         tokens.AccessToken,
		"refresh_token": tokens.RefreshToken,
		"expires_in":    tokens.ExpiresIn,
		"token_type":    tokens.TokenType,
		"user":          user.ToResponse(),
	})
}

// RefreshToken 刷新 Token
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	response.Error(c, http.StatusNotImplemented, "刷新 Token 功能尚未实现")
}
