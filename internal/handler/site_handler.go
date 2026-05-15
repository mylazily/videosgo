package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/service"
	"github.com/mylazily/videosgo/pkg/response"
)

// SiteHandler 站群管理处理器
type SiteHandler struct {
	svc *service.SiteService
}

// NewSiteHandler 创建站群管理处理器
func NewSiteHandler(svc *service.SiteService) *SiteHandler {
	return &SiteHandler{svc: svc}
}

// ListSites 获取域名列表
// GET /api/v1/admin/sites
func (h *SiteHandler) ListSites(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	cluster := c.Query("cluster")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	sites, total, err := h.svc.ListSites(page, pageSize, cluster)
	if err != nil {
		response.InternalError(c, "获取域名列表失败")
		return
	}

	response.SuccessPage(c, sites, total, page, pageSize)
}

// CreateSite 创建域名
// POST /api/v1/admin/sites
func (h *SiteHandler) CreateSite(c *gin.Context) {
	var req model.SiteDomain
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if req.Domain == "" {
		response.BadRequest(c, "域名不能为空")
		return
	}
	if req.Cluster != "A" && req.Cluster != "B" {
		response.BadRequest(c, "集群必须为 A 或 B")
		return
	}

	if err := h.svc.CreateSite(&req); err != nil {
		response.InternalError(c, "创建域名失败")
		return
	}

	response.Success(c, req)
}

// UpdateSite 更新域名
// PUT /api/v1/admin/sites/:id
func (h *SiteHandler) UpdateSite(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少域名 ID")
		return
	}

	site, err := h.svc.GetSite(id)
	if err != nil {
		response.NotFound(c, "域名不存在")
		return
	}

	var req model.SiteDomain
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 保留不可修改的字段
	req.ID = site.ID
	req.CreatedAt = site.CreatedAt

	if err := h.svc.UpdateSite(&req); err != nil {
		response.InternalError(c, "更新域名失败")
		return
	}

	response.Success(c, req)
}

// DeleteSite 删除域名
// DELETE /api/v1/admin/sites/:id
func (h *SiteHandler) DeleteSite(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少域名 ID")
		return
	}

	if err := h.svc.DeleteSite(id); err != nil {
		response.InternalError(c, "删除域名失败")
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// HealthCheck 手动健康检查
// POST /api/v1/admin/sites/:id/health-check
func (h *SiteHandler) HealthCheck(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少域名 ID")
		return
	}

	if err := h.svc.HealthCheckSingle(id); err != nil {
		response.InternalError(c, "健康检查失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "健康检查已触发", nil)
}

// GetLinkAudit 获取交叉链接审计
// GET /api/v1/admin/sites/audit
func (h *SiteHandler) GetLinkAudit(c *gin.Context) {
	audits, err := h.svc.GetCrossClusterLinks()
	if err != nil {
		response.InternalError(c, "获取审计记录失败")
		return
	}

	response.Success(c, audits)
}
