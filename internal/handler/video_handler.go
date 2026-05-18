package handler

import (
	"strconv"

	"videosgo/internal/service"
	"videosgo/pkg/response"

	"github.com/gin-gonic/gin"
)

// VideoHandler 视频处理器
type VideoHandler struct {
	videoService *service.VideoService
}

// NewVideoHandler 创建视频处理器
func NewVideoHandler(videoService *service.VideoService) *VideoHandler {
	return &VideoHandler{videoService: videoService}
}

// ListVideos 获取视频列表
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
	videos, err := h.videoService.ListVideos(offset, pageSize)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, videos)
}

// GetVideo 获取视频详情
func (h *VideoHandler) GetVideo(c *gin.Context) {
	id := c.Param("id")
	video, err := h.videoService.GetVideo(id)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	if video == nil {
		response.NotFound(c, "视频不存在")
		return
	}

	response.Success(c, video)
}

// FavoriteVideo 收藏视频
func (h *VideoHandler) FavoriteVideo(c *gin.Context) {
	// TODO: 实现收藏逻辑
	response.SuccessWithMessage(c, "收藏成功", nil)
}

// GetUserVideos 获取用户视频
func (h *VideoHandler) GetUserVideos(c *gin.Context) {
	// TODO: 实现获取用户视频逻辑
	response.Success(c, []interface{}{})
}
