package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/service"
	"github.com/mylazily/videosgo/pkg/response"
)

// TagHandler 标签处理器
type TagHandler struct {
	svc *service.TagService
}

// NewTagHandler 创建标签处理器
func NewTagHandler(svc *service.TagService) *TagHandler {
	return &TagHandler{svc: svc}
}

// ListTags 获取标签列表
// GET /api/v1/tags?sort=trending&keyword=xxx&page=1&page_size=20
func (h *TagHandler) ListTags(c *gin.Context) {
	sort := c.DefaultQuery("sort", "default")
	keyword := c.Query("keyword")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	// 搜索模式
	if keyword != "" {
		tags, err := h.svc.SearchTags(keyword)
		if err != nil {
			response.InternalError(c, "搜索标签失败")
			return
		}
		response.Success(c, tags)
		return
	}

	// 热门模式
	if sort == "trending" {
		limit := pageSize
		tags, err := h.svc.GetTrendingTags(limit)
		if err != nil {
			response.InternalError(c, "获取热门标签失败")
			return
		}
		response.Success(c, tags)
		return
	}

	// 默认列表模式
	tags, total, err := h.svc.ListTags(page, pageSize)
	if err != nil {
		response.InternalError(c, "获取标签列表失败")
		return
	}

	response.SuccessPage(c, tags, total, page, pageSize)
}

// GetTagBySlug 获取标签详情
// GET /api/v1/tags/:slug
func (h *TagHandler) GetTagBySlug(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.BadRequest(c, "标签标识不能为空")
		return
	}

	tag, err := h.svc.GetTagBySlug(slug)
	if err != nil {
		response.NotFound(c, "标签不存在")
		return
	}

	response.Success(c, tag)
}

// GetTagVideos 获取标签下的视频
// GET /api/v1/tags/:slug/videos?page=1&page_size=20
func (h *TagHandler) GetTagVideos(c *gin.Context) {
	slug := c.Param("slug")
	if slug == "" {
		response.BadRequest(c, "标签标识不能为空")
		return
	}

	// 先获取标签
	tag, err := h.svc.GetTagBySlug(slug)
	if err != nil {
		response.NotFound(c, "标签不存在")
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	videos, total, err := h.svc.GetVideosByTag(tag.ID, page, pageSize)
	if err != nil {
		response.InternalError(c, "获取标签视频失败")
		return
	}

	response.SuccessPage(c, videos, total, page, pageSize)
}

// GetVideoTags 获取视频的标签
// GET /api/v1/videos/:id/tags
func (h *TagHandler) GetVideoTags(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的视频 ID")
		return
	}

	tags, err := h.svc.GetVideoTags(id)
	if err != nil {
		response.InternalError(c, "获取视频标签失败")
		return
	}

	response.Success(c, tags)
}
