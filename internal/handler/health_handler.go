package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mylazily/videosgo/pkg/response"
)

// HealthHandler 健康检查处理器
type HealthHandler struct{}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// Ping 健康检查
// GET /api/v1/ping
func (h *HealthHandler) Ping(c *gin.Context) {
	response.Success(c, gin.H{
		"status":  "ok",
		"message": "pong",
	})
}

// Health 详细健康检查
// GET /api/v1/health
func (h *HealthHandler) Health(c *gin.Context) {
	response.Success(c, gin.H{
		"status": "healthy",
	})
}
