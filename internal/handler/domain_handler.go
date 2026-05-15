package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mylazily/videosgo/internal/service"
	"github.com/mylazily/videosgo/pkg/response"
)

// DomainHandler 域名轮询处理器
type DomainHandler struct {
	svc *service.DomainRotationService
}

// NewDomainHandler 创建域名轮询处理器
func NewDomainHandler(svc *service.DomainRotationService) *DomainHandler {
	return &DomainHandler{svc: svc}
}

// GetActiveDomain 获取当前活跃域名
// GET /api/v1/domain/active
func (h *DomainHandler) GetActiveDomain(c *gin.Context) {
	region := c.Query("region")
	domain, err := h.svc.GetActiveDomain(region)
	if err != nil {
		response.NotFound(c, "未找到活跃域名")
		return
	}

	response.Success(c, gin.H{
		"domain": *domain,
		"region": region,
	})
}

// GetDomainList 域名列表及状态
// GET /api/v1/domain/list
func (h *DomainHandler) GetDomainList(c *gin.Context) {
	data, err := h.svc.GetDomainList()
	if err != nil {
		response.InternalError(c, "获取域名列表失败")
		return
	}
	response.Success(c, data)
}

// SwitchDomain 手动切换域名（管理员）
// POST /api/v1/domain/switch
func (h *DomainHandler) SwitchDomain(c *gin.Context) {
	var req struct {
		Domain string `json:"domain" binding:"required"`
		Region string `json:"region"`
		Reason string `json:"reason" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.SwitchDomain(req.Domain, req.Region, req.Reason); err != nil {
		response.InternalError(c, "切换域名失败")
		return
	}

	response.SuccessWithMessage(c, "域名切换成功", nil)
}

// GetSwitchHistory 切换历史
// GET /api/v1/domain/history
func (h *DomainHandler) GetSwitchHistory(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "20")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 20
	}

	data, err := h.svc.GetSwitchHistory(limit)
	if err != nil {
		response.InternalError(c, "获取切换历史失败")
		return
	}
	response.Success(c, data)
}
