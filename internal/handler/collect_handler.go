// Package handler HTTP 请求处理器
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"videosgo/internal/model"
	"videosgo/internal/service"
	"videosgo/pkg/response"
)

// CollectHandler 采集源处理器
type CollectHandler struct {
	svc *service.CollectService
}

// NewCollectHandler 创建采集源处理器
func NewCollectHandler(svc *service.CollectService) *CollectHandler {
	return &CollectHandler{svc: svc}
}

// CreateSource 创建采集源
// POST /api/v1/admin/collect/sources
func (h *CollectHandler) CreateSource(c *gin.Context) {
	var source model.CollectSource
	if err := c.ShouldBindJSON(&source); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.svc.CreateSource(&source); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, source)
}

// UpdateSource 更新采集源
// PUT /api/v1/admin/collect/sources/:id
func (h *CollectHandler) UpdateSource(c *gin.Context) {
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的 ID")
		return
	}

	var source model.CollectSource
	if err := c.ShouldBindJSON(&source); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	source.ID = parsedID

	if err := h.svc.UpdateSource(&source); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, source)
}

// DeleteSource 删除采集源
// DELETE /api/v1/admin/collect/sources/:id
func (h *CollectHandler) DeleteSource(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.DeleteSource(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// GetSource 获取采集源详情
// GET /api/v1/admin/collect/sources/:id
func (h *CollectHandler) GetSource(c *gin.Context) {
	id := c.Param("id")

	source, err := h.svc.GetSource(id)
	if err != nil {
		response.NotFound(c, "采集源不存在")
		return
	}

	response.Success(c, source)
}

// ListSources 获取采集源列表
// GET /api/v1/admin/collect/sources
func (h *CollectHandler) ListSources(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	sources, total, err := h.svc.ListSources(page, pageSize)
	if err != nil {
		response.InternalError(c, "获取采集源列表失败")
		return
	}

	response.SuccessPage(c, sources, total, page, pageSize)
}

// TriggerCollect 触发采集
// POST /api/v1/admin/collect/sources/:id/trigger
func (h *CollectHandler) TriggerCollect(c *gin.Context) {
	id := c.Param("id")

	collectType := c.DefaultQuery("type", "full")

	if err := h.svc.TriggerCollect(id, collectType); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "采集任务已提交", nil)
}

// ListLogs 获取采集日志
// GET /api/v1/admin/collect/logs
func (h *CollectHandler) ListLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := h.svc.ListLogs(page, pageSize)
	if err != nil {
		response.InternalError(c, "获取采集日志失败")
		return
	}

	response.SuccessPage(c, logs, total, page, pageSize)
}
