package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/service"
	"github.com/mylazily/videosgo/pkg/response"
)

// ShortVideoHandler 短视频处理器
type ShortVideoHandler struct {
	svc *service.ShortVideoService
}

// NewShortVideoHandler 创建短视频处理器
func NewShortVideoHandler(svc *service.ShortVideoService) *ShortVideoHandler {
	return &ShortVideoHandler{svc: svc}
}

// ListShorts 获取短视频列表
// GET /api/v1/shorts?sort=popular/latest/random&page=1&size=20
func (h *ShortVideoHandler) ListShorts(c *gin.Context) {
	sort := c.DefaultQuery("sort", "latest")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 随机模式
	if sort == "random" {
		shorts, err := h.svc.GetRandom(pageSize)
		if err != nil {
			response.InternalError(c, "获取短视频失败")
			return
		}
		response.Success(c, shorts)
		return
	}

	shorts, total, err := h.svc.ListShortVideos(page, pageSize, sort)
	if err != nil {
		response.InternalError(c, "获取短视频列表失败")
		return
	}

	response.SuccessPage(c, shorts, total, page, pageSize)
}

// GetShort 获取短视频详情
// GET /api/v1/shorts/:id
func (h *ShortVideoHandler) GetShort(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的短视频 ID")
		return
	}

	sv, err := h.svc.GetShortVideo(id)
	if err != nil {
		response.NotFound(c, "短视频不存在")
		return
	}

	response.Success(c, sv)
}

// IncrementView 增加播放量
// POST /api/v1/shorts/:id/view
func (h *ShortVideoHandler) IncrementView(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的短视频 ID")
		return
	}

	if err := h.svc.IncrementView(id); err != nil {
		response.InternalError(c, "操作失败")
		return
	}

	response.SuccessWithMessage(c, "操作成功", nil)
}

// IncrementLike 点赞
// POST /api/v1/shorts/:id/like
func (h *ShortVideoHandler) IncrementLike(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的短视频 ID")
		return
	}

	if err := h.svc.IncrementLike(id); err != nil {
		response.InternalError(c, "操作失败")
		return
	}

	response.SuccessWithMessage(c, "点赞成功", nil)
}

// GetRandom 获取随机发现
// GET /api/v1/shorts/random?limit=10
func (h *ShortVideoHandler) GetRandom(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	shorts, err := h.svc.GetRandom(limit)
	if err != nil {
		response.InternalError(c, "获取随机短视频失败")
		return
	}

	response.Success(c, shorts)
}
