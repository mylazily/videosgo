package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/service"
	"github.com/mylazily/videosgo/pkg/response"
)

// PaymentHandler 支付处理器
type PaymentHandler struct {
	svc *service.PaymentService
}

// NewPaymentHandler 创建支付处理器
func NewPaymentHandler(svc *service.PaymentService) *PaymentHandler {
	return &PaymentHandler{svc: svc}
}

// CreateOrder 创建支付订单
// POST /api/v1/payment/create
func (h *PaymentHandler) CreateOrder(c *gin.Context) {
	var req struct {
		FingerprintID string  `json:"fingerprint_id"`
		ChannelID     string  `json:"channel_id" binding:"required"`
		ProductType   string  `json:"product_type" binding:"required"`
		ProductID     string  `json:"product_id" binding:"required"`
		ProductName   string  `json:"product_name" binding:"required"`
		Amount        float64 `json:"amount" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	var fingerprintID *uuid.UUID
	if req.FingerprintID != "" {
		id, err := uuid.Parse(req.FingerprintID)
		if err != nil {
			response.BadRequest(c, "无效的设备指纹 ID")
			return
		}
		fingerprintID = &id
	}

	channelID, err := uuid.Parse(req.ChannelID)
	if err != nil {
		response.BadRequest(c, "无效的支付渠道 ID")
		return
	}

	order, err := h.svc.CreateOrder(fingerprintID, channelID, req.ProductType, req.ProductID, req.ProductName, req.Amount)
	if err != nil {
		response.InternalError(c, "创建订单失败")
		return
	}

	response.Success(c, order)
}

// GetOrder 查询订单状态
// GET /api/v1/payment/:orderNo
func (h *PaymentHandler) GetOrder(c *gin.Context) {
	orderNo := c.Param("orderNo")
	if orderNo == "" {
		response.BadRequest(c, "订单号不能为空")
		return
	}

	order, err := h.svc.GetOrder(orderNo)
	if err != nil {
		response.NotFound(c, "订单不存在")
		return
	}

	response.Success(c, order)
}

// VerifyVIP 验证 VIP 权限
// POST /api/v1/payment/verify
func (h *PaymentHandler) VerifyVIP(c *gin.Context) {
	var req struct {
		FingerprintID string `json:"fingerprint_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	fingerprintID, err := uuid.Parse(req.FingerprintID)
	if err != nil {
		response.BadRequest(c, "无效的设备指纹 ID")
		return
	}

	sub, err := h.svc.VerifyVIPAccess(fingerprintID)
	if err != nil {
		response.Success(c, gin.H{
			"is_vip": false,
			"message": err.Error(),
		})
		return
	}

	response.Success(c, gin.H{
		"is_vip":     true,
		"plan_type":  sub.PlanType,
		"expires_at": sub.ExpiresAt,
	})
}

// ListChannels 支付渠道列表
// GET /api/v1/payment/channels
func (h *PaymentHandler) ListChannels(c *gin.Context) {
	channels, err := h.svc.GetChannels()
	if err != nil {
		response.InternalError(c, "获取支付渠道失败")
		return
	}
	response.Success(c, channels)
}

// GetVIPStatus VIP 状态
// GET /api/v1/payment/vip/status
func (h *PaymentHandler) GetVIPStatus(c *gin.Context) {
	fingerprintIDStr := c.Query("fingerprint_id")
	if fingerprintIDStr == "" {
		response.BadRequest(c, "设备指纹 ID 不能为空")
		return
	}

	fingerprintID, err := uuid.Parse(fingerprintIDStr)
	if err != nil {
		response.BadRequest(c, "无效的设备指纹 ID")
		return
	}

	sub, err := h.svc.GetVIPStatus(fingerprintID)
	if err != nil {
		response.Success(c, gin.H{
			"is_vip": false,
		})
		return
	}

	response.Success(c, gin.H{
		"is_vip":     true,
		"plan_type":  sub.PlanType,
		"start_at":   sub.StartAt,
		"expires_at": sub.ExpiresAt,
		"auto_renew": sub.AutoRenew,
	})
}
