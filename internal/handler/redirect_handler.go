package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"videosgo/internal/model"
	"videosgo/internal/service"
	"videosgo/pkg/response"
)

// RedirectHandler 301 重定向处理器
type RedirectHandler struct {
	svc *service.RedirectService
}

// NewRedirectHandler 创建重定向处理器
func NewRedirectHandler(svc *service.RedirectService) *RedirectHandler {
	return &RedirectHandler{svc: svc}
}

// ListRules 获取规则列表
// GET /api/v1/admin/redirects
func (h *RedirectHandler) ListRules(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	rules, total, err := h.svc.ListRules(page, pageSize)
	if err != nil {
		response.InternalError(c, "获取规则列表失败")
		return
	}

	response.SuccessPage(c, rules, total, page, pageSize)
}

// CreateRule 创建规则
// POST /api/v1/admin/redirects
func (h *RedirectHandler) CreateRule(c *gin.Context) {
	var req model.RedirectRule
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	if req.SourceDomain == "" {
		response.BadRequest(c, "来源域名不能为空")
		return
	}
	if req.TargetURL == "" {
		response.BadRequest(c, "目标 URL 不能为空")
		return
	}

	if err := h.svc.CreateRule(&req); err != nil {
		response.InternalError(c, "创建规则失败")
		return
	}

	response.Success(c, req)
}

// UpdateRule 更新规则
// PUT /api/v1/admin/redirects/:id
func (h *RedirectHandler) UpdateRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少规则 ID")
		return
	}

	rule, err := h.svc.GetRule(id)
	if err != nil {
		response.NotFound(c, "规则不存在")
		return
	}

	var req model.RedirectRule
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误")
		return
	}

	// 保留不可修改的字段
	req.ID = rule.ID
	req.CreatedAt = rule.CreatedAt

	if err := h.svc.UpdateRule(&req); err != nil {
		response.InternalError(c, "更新规则失败")
		return
	}

	response.Success(c, req)
}

// DeleteRule 删除规则
// DELETE /api/v1/admin/redirects/:id
func (h *RedirectHandler) DeleteRule(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少规则 ID")
		return
	}

	if err := h.svc.DeleteRule(id); err != nil {
		response.InternalError(c, "删除规则失败")
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// GetHitLogs 获取命中日志
// GET /api/v1/admin/redirects/:id/logs
func (h *RedirectHandler) GetHitLogs(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "缺少规则 ID")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit < 1 || limit > 200 {
		limit = 50
	}

	logs, err := h.svc.GetHitLogs(id, limit)
	if err != nil {
		response.InternalError(c, "获取命中日志失败")
		return
	}

	response.Success(c, logs)
}
