package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/service"
	"github.com/mylazily/videosgo/pkg/response"
)

// TGHandler TG Bot 处理器
type TGHandler struct {
	svc *service.TGService
}

// NewTGHandler 创建 TG Bot 处理器
func NewTGHandler(svc *service.TGService) *TGHandler {
	return &TGHandler{svc: svc}
}

// Webhook TG Webhook 接收
// POST /api/v1/tg/webhook
func (h *TGHandler) Webhook(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// TODO: 解析 TG Webhook 数据并处理
	// 目前仅记录接收
	response.Success(c, gin.H{
		"received": true,
		"update_id": body["update_id"],
	})
}

// ListChannels 频道列表
// GET /api/v1/tg/channels
func (h *TGHandler) ListChannels(c *gin.Context) {
	channels, err := h.svc.GetChannels()
	if err != nil {
		response.InternalError(c, "获取频道列表失败")
		return
	}
	response.Success(c, channels)
}

// Broadcast 手动群发（管理员）
// POST /api/v1/tg/broadcast
func (h *TGHandler) Broadcast(c *gin.Context) {
	var req struct {
		VideoID string `json:"video_id" binding:"required"`
		Text    string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	videoID, err := uuid.Parse(req.VideoID)
	if err != nil {
		response.BadRequest(c, "无效的视频 ID")
		return
	}

	// 异步群发
	h.svc.BroadcastToAllChannels(videoID, req.Text)

	response.SuccessWithMessage(c, "群发任务已提交", nil)
}

// RegisterMiniAppSession Mini App 会话注册
// POST /api/v1/tg/miniapp/session
func (h *TGHandler) RegisterMiniAppSession(c *gin.Context) {
	var req struct {
		TGUserID      int64   `json:"tg_user_id" binding:"required"`
		TGUsername    string  `json:"tg_username"`
		TGLanguage    string  `json:"tg_language"`
		FingerprintID *string `json:"fingerprint_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var fingerprintID *uuid.UUID
	if req.FingerprintID != nil && *req.FingerprintID != "" {
		id, err := uuid.Parse(*req.FingerprintID)
		if err != nil {
			response.BadRequest(c, "无效的设备指纹 ID")
			return
		}
		fingerprintID = &id
	}

	session, err := h.svc.RegisterMiniAppSession(req.TGUserID, req.TGUsername, req.TGLanguage, fingerprintID)
	if err != nil {
		response.InternalError(c, "注册会话失败")
		return
	}

	response.Success(c, session)
}

// GetMiniAppStats Mini App 统计
// GET /api/v1/tg/miniapp/stats
func (h *TGHandler) GetMiniAppStats(c *gin.Context) {
	totalSessions, totalWatchTime, err := h.svc.GetMiniAppStats()
	if err != nil {
		response.InternalError(c, "获取统计失败")
		return
	}

	response.Success(c, gin.H{
		"total_sessions":  totalSessions,
		"total_watch_time": totalWatchTime,
	})
}
