package handler

import (
	"net/http"

	"videosgo/internal/service"

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
	videos, err := h.videoService.ListVideos(0, 20)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    videos,
	})
}

// GetVideo 获取视频详情
func (h *VideoHandler) GetVideo(c *gin.Context) {
	id := c.Param("id")
	video, err := h.videoService.GetVideo(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    500,
			"message": err.Error(),
		})
		return
	}

	if video == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"code":    404,
			"message": "视频不存在",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    video,
	})
}

// FavoriteVideo 收藏视频
func (h *VideoHandler) FavoriteVideo(c *gin.Context) {
	// TODO: 实现收藏逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "收藏成功",
	})
}

// GetUserVideos 获取用户视频
func (h *VideoHandler) GetUserVideos(c *gin.Context) {
	// TODO: 实现获取用户视频逻辑
	c.JSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "获取成功",
		"data":    []interface{}{},
	})
}
