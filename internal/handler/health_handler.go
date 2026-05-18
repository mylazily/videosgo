package handler

import (
	"context"
	"database/sql"

	"videosgo/pkg/response"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// HealthHandler 健康检查处理器
type HealthHandler struct {
	db    *sql.DB
	redis *redis.Client
}

// NewHealthHandler 创建健康检查处理器
func NewHealthHandler(db *sql.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

// Ping 简单的 ping 检查
func (h *HealthHandler) Ping(c *gin.Context) {
	response.SuccessWithMessage(c, "pong", nil)
}

// Health 健康检查（兼容前端 apiConfigStore 格式）
func (h *HealthHandler) Health(c *gin.Context) {
	status := "healthy"
	checks := make(map[string]string)

	// 检查数据库
	if h.db != nil {
		if err := h.db.Ping(); err != nil {
			status = "degraded"
			checks["database"] = "error"
		} else {
			checks["database"] = "ok"
		}
	}

	// 检查 Redis
	if h.redis != nil {
		if err := h.redis.Ping(context.Background()).Err(); err != nil {
			status = "degraded"
			checks["redis"] = "error"
		} else {
			checks["redis"] = "ok"
		}
	}

	response.Success(c, gin.H{
		"status":  status,
		"service": "videosgo",
		"version": "1.0.0",
		"checks":  checks,
	})
}
