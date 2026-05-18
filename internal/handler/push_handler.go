package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"videosgo/internal/service"
	"videosgo/pkg/response"
)

// PushHandler 推送处理器
type PushHandler struct {
	svc *service.PushService
}

// NewPushHandler 创建推送处理器
func NewPushHandler(svc *service.PushService) *PushHandler {
	return &PushHandler{svc: svc}
}

// Subscribe 订阅推送
// POST /api/v1/push/subscribe
func (h *PushHandler) Subscribe(c *gin.Context) {
	var req struct {
		FingerprintID string `json:"fingerprint_id" binding:"required"`
		Endpoint      string `json:"endpoint" binding:"required"`
		Keys          struct {
			P256DH string `json:"p256dh"`
			Auth   string `json:"auth"`
		} `json:"keys"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	sub, err := h.svc.Subscribe(
		req.FingerprintID,
		req.Endpoint,
		req.Keys.P256DH,
		req.Keys.Auth,
		c.GetHeader("User-Agent"),
	)
	if err != nil {
		response.InternalError(c, "订阅失败")
		return
	}

	response.Success(c, sub)
}

// Unsubscribe 取消订阅
// DELETE /api/v1/push/subscribe
func (h *PushHandler) Unsubscribe(c *gin.Context) {
	var req struct {
		SubscriptionID string `json:"subscription_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if err := h.svc.Unsubscribe(req.SubscriptionID); err != nil {
		response.InternalError(c, "取消订阅失败")
		return
	}

	response.SuccessWithMessage(c, "取消订阅成功", nil)
}

// SendNotification 发送推送（管理员）
// POST /api/v1/push/send
func (h *PushHandler) SendNotification(c *gin.Context) {
	var req struct {
		Title         string `json:"title" binding:"required"`
		Body          string `json:"body" binding:"required"`
		Icon          string `json:"icon"`
		Link          string `json:"link"`
		Tag           string `json:"tag"`
		TargetType    string `json:"target_type"`
		TargetVideoID string `json:"target_video_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	notif, err := h.svc.SendNotification(
		req.Title, req.Body, req.Icon, req.Link,
		req.Tag, req.TargetType, req.TargetVideoID,
	)
	if err != nil {
		response.InternalError(c, "发送推送失败")
		return
	}

	response.Success(c, notif)
}

// GetStats 获取推送统计
// GET /api/v1/push/stats
func (h *PushHandler) GetStats(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	stats, err := h.svc.GetNotificationStats(limit)
	if err != nil {
		response.InternalError(c, "获取推送统计失败")
		return
	}

	response.Success(c, stats)
}
