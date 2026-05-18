package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"videosgo/internal/service"
	"videosgo/pkg/response"
)

// RankHandler 排行榜处理器
type RankHandler struct {
	svc *service.RankService
}

// NewRankHandler 创建排行榜处理器
func NewRankHandler(svc *service.RankService) *RankHandler {
	return &RankHandler{svc: svc}
}

// GetDailyRank 获取日排行榜
// GET /api/v1/rank/daily
func (h *RankHandler) GetDailyRank(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	ranks, err := h.svc.GetDailyRank(limit)
	if err != nil {
		response.InternalError(c, "获取日排行榜失败")
		return
	}

	response.Success(c, ranks)
}

// GetWeeklyRank 获取周排行榜
// GET /api/v1/rank/weekly
func (h *RankHandler) GetWeeklyRank(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	ranks, err := h.svc.GetWeeklyRank(limit)
	if err != nil {
		response.InternalError(c, "获取周排行榜失败")
		return
	}

	response.Success(c, ranks)
}

// GetMonthlyRank 获取月排行榜
// GET /api/v1/rank/monthly
func (h *RankHandler) GetMonthlyRank(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	ranks, err := h.svc.GetMonthlyRank(limit)
	if err != nil {
		response.InternalError(c, "获取月排行榜失败")
		return
	}

	response.Success(c, ranks)
}

// GetCategoryRank 获取分类排行榜
// GET /api/v1/rank/category/:category
func (h *RankHandler) GetCategoryRank(c *gin.Context) {
	category := c.Param("category")
	rankType := c.DefaultQuery("type", "daily")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit < 1 || limit > 100 {
		limit = 20
	}

	ranks, err := h.svc.GetCategoryRank(category, rankType, limit)
	if err != nil {
		response.InternalError(c, "获取分类排行榜失败")
		return
	}

	response.Success(c, ranks)
}
