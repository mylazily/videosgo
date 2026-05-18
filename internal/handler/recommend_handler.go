package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"videosgo/internal/service"
	"videosgo/pkg/response"
)

// RecommendHandler 推荐处理器
type RecommendHandler struct {
	svc *service.RecommendService
}

// NewRecommendHandler 创建推荐处理器
func NewRecommendHandler(svc *service.RecommendService) *RecommendHandler {
	return &RecommendHandler{svc: svc}
}

// GetRelatedVideos 获取相关推荐
// GET /api/v1/videos/:id/related?limit=10
func (h *RecommendHandler) GetRelatedVideos(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的视频 ID")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	videos, err := h.svc.GetRelatedVideos(id, limit)
	if err != nil {
		response.InternalError(c, "获取相关推荐失败")
		return
	}

	response.Success(c, videos)
}

// GetPersonalizedRecommendations 获取个性化推荐
// GET /api/v1/recommendations?fingerprint_id=xxx&limit=10
func (h *RecommendHandler) GetPersonalizedRecommendations(c *gin.Context) {
	fingerprintID := c.Query("fingerprint_id")
	if fingerprintID == "" {
		response.BadRequest(c, "设备指纹 ID 不能为空")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	videos, err := h.svc.GetPersonalizedRecommendations(fingerprintID, limit)
	if err != nil {
		response.InternalError(c, "获取个性化推荐失败")
		return
	}

	response.Success(c, videos)
}
