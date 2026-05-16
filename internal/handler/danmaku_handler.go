package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/service"
	"github.com/mylazily/videosgo/pkg/response"
)

// DanmakuHandler 弹幕处理器
type DanmakuHandler struct {
	svc *service.DanmakuService
}

// NewDanmakuHandler 创建弹幕处理器
func NewDanmakuHandler(svc *service.DanmakuService) *DanmakuHandler {
	return &DanmakuHandler{svc: svc}
}

// CreateDanmaku 创建弹幕
// POST /api/v1/videos/:id/episodes/:ep_id/danmaku
func (h *DanmakuHandler) CreateDanmaku(c *gin.Context) {
	videoIDStr := c.Param("id")
	parsedVideoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		response.BadRequest(c, "无效的视频 ID")
		return
	}

	episodeIDStr := c.Param("ep_id")
	parsedEpisodeID, err := uuid.Parse(episodeIDStr)
	if err != nil {
		response.BadRequest(c, "无效的剧集 ID")
		return
	}

	var req struct {
		Time    float64 `json:"time" binding:"required"`
		Type    int     `json:"type"`
		Color   string  `json:"color"`
		Content string  `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")
	parsedUserID, err := uuid.Parse(userID.(string))
	if err != nil {
		response.BadRequest(c, "无效的用户 ID")
		return
	}

	danmaku := &model.Danmaku{
		VideoID:   parsedVideoID,
		EpisodeID: parsedEpisodeID,
		UserID:    parsedUserID,
		Time: fmt.Sprintf("%.0f", req.Time),
		Type: fmt.Sprintf("%d", req.Type),
		Color:     req.Color,
		Content:   req.Content,
	}

	if err := h.svc.CreateDanmaku(danmaku); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, danmaku)
}

// GetDanmakus 获取剧集弹幕
// GET /api/v1/videos/:id/episodes/:ep_id/danmaku
func (h *DanmakuHandler) GetDanmakus(c *gin.Context) {
	episodeIDStr := c.Param("ep_id")
	parsedEpisodeID, err := uuid.Parse(episodeIDStr)
	if err != nil {
		response.BadRequest(c, "无效的剧集 ID")
		return
	}

	danmakus, err := h.svc.GetDanmakusByEpisode(parsedEpisodeID)
	if err != nil {
		response.InternalError(c, "获取弹幕失败")
		return
	}

	response.Success(c, danmakus)
}

// GetVideoDanmakus 获取视频所有弹幕
// GET /api/v1/videos/:id/danmaku
func (h *DanmakuHandler) GetVideoDanmakus(c *gin.Context) {
	videoIDStr := c.Param("id")
	parsedVideoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		response.BadRequest(c, "无效的视频 ID")
		return
	}

	danmakus, err := h.svc.GetDanmakusByVideo(parsedVideoID)
	if err != nil {
		response.InternalError(c, "获取弹幕失败")
		return
	}

	response.Success(c, danmakus)
}
