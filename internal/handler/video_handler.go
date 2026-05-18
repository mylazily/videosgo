package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"videosgo/internal/model"
	"videosgo/internal/service"
	"videosgo/pkg/response"
)

// VideoHandler 视频处理器
type VideoHandler struct {
	videoSvc   *service.VideoService
	commentSvc *service.CommentService
	danmakuSvc *service.DanmakuService
	tagSvc    *service.TagService
}

// NewVideoHandler 创建视频处理器
func NewVideoHandler(
	videoSvc *service.VideoService,
	commentSvc *service.CommentService,
	danmakuSvc *service.DanmakuService,
	tagSvc *service.TagService,
) *VideoHandler {
	return &VideoHandler{
		videoSvc:   videoSvc,
		commentSvc: commentSvc,
		danmakuSvc: danmakuSvc,
		tagSvc:    tagSvc,
	}
}

// ListVideos 获取视频列表
// GET /api/v1/videos?page=1&page_size=20&category=xxx&sort=latest
func (h *VideoHandler) ListVideos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	videos, err := h.videoSvc.ListVideos(offset, pageSize)
	if err != nil {
		response.InternalError(c, "获取视频列表失败")
		return
	}

	response.Success(c, gin.H{
		"list":      videos,
		"page":      page,
		"page_size": pageSize,
	})
}

// GetVideo 获取视频详情
// GET /api/v1/videos/:id
func (h *VideoHandler) GetVideo(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "视频 ID 不能为空")
		return
	}

	video, err := h.videoSvc.GetVideo(id)
	if err != nil {
		response.NotFound(c, "视频不存在")
		return
	}

	if video == nil {
		response.NotFound(c, "视频不存在")
		return
	}

	response.Success(c, video)
}

// CreateVideo 创建视频（管理员）
// POST /api/v1/admin/videos
func (h *VideoHandler) CreateVideo(c *gin.Context) {
	var video model.Video
	if err := c.ShouldBindJSON(&video); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	if err := h.videoSvc.CreateVideo(&video); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, video)
}

// UpdateVideo 更新视频（管理员）
// PUT /api/v1/admin/videos/:id
func (h *VideoHandler) UpdateVideo(c *gin.Context) {
	idStr := c.Param("id")
	parsedID, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的视频 ID")
		return
	}

	var video model.Video
	if err := c.ShouldBindJSON(&video); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}
	video.ID = parsedID

	if err := h.videoSvc.UpdateVideo(&video); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, video)
}

// DeleteVideo 删除视频（管理员）
// DELETE /api/v1/admin/videos/:id
func (h *VideoHandler) DeleteVideo(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		response.BadRequest(c, "视频 ID 不能为空")
		return
	}

	if err := h.videoSvc.DeleteVideo(id); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}
