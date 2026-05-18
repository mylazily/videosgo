package handler

import (
	"database/sql"
	"net/http"

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
	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "pong",
	})
}

// Health 健康检查（兼容前端 apiConfigStore 格式）
func (h *HealthHandler) Health(c *gin.Context) {
	status := "healthy"
	checks := gin.H{}

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
		if err := h.redis.Ping(c).Err(); err != nil {
			status = "degraded"
			checks["redis"] = "error"
		} else {
			checks["redis"] = "ok"
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    0,
		"message": "success",
		"data": gin.H{
			"status":  status,
			"service": "videosgo",
			"version": "1.0.0",
			"checks":  checks,
		},
	})
}
