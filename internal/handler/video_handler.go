package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/service"
	"github.com/mylazily/videosgo/pkg/response"
)

// VideoHandler 视频处理器
type VideoHandler struct {
	svc     *service.VideoService
	monitor *service.StationMonitor // 资源站监控（可以为 nil）
}

// NewVideoHandler 创建视频处理器
func NewVideoHandler(svc *service.VideoService) *VideoHandler {
	return &VideoHandler{svc: svc}
}

// SetStationMonitor 设置资源站监控（可选注入）
func (h *VideoHandler) SetStationMonitor(monitor *service.StationMonitor) {
	h.monitor = monitor
}

// GetVideo 获取视频详情
// GET /api/v1/videos/:id
// 返回增强数据：play_lines、domain_pool、shared_path、source_count
func (h *VideoHandler) GetVideo(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的视频 ID")
		return
	}

	video, err := h.svc.GetVideo(id)
	if err != nil {
		response.NotFound(c, "视频不存在")
		return
	}

	// 构建增强响应
	result := gin.H{
		"id":           video.ID,
		"title":        video.Title,
		"sub_title":    video.SubTitle,
		"cover":        video.CoverURL,
		"description":  video.Description,
		"category_id":  video.Category,
		"category":     video.Category,
		"year":         video.Year,
		"area":         video.Area,
		"director":     video.Director,
		"actors":       video.Actors,
		"tags":         video.Tags,
		"remarks":      video.Remarks,
		"play_links":   video.PlayLinks,
		"status":       video.Status,
		"source_id":    video.SourceID,
		"view_count":   video.ViewCount,
		"like_count":   video.LikeCount,
		"score":        video.Score,
		"created_at":   video.CreatedAt,
		"updated_at":   video.UpdatedAt,
		"episodes":     video.Episodes,
		// 增强字段
		"clean_title":  video.CleanTitle,
		"play_lines":   h.enrichPlayLines(video.PlayLines),
		"domain_pool":  video.DomainPool,
		"shared_path":  video.SharedPath,
		"source_count": video.SourceCount,
	}

	// 如果有域名池和共享路径，额外返回备用 URL 列表
	if video.SharedPath != "" && len(video.DomainPool) > 0 {
		alternateURLs := make([]string, 0, len(video.DomainPool))
		for _, domain := range video.DomainPool {
			alternateURLs = append(alternateURLs, "https://"+domain+video.SharedPath)
		}
		result["alternate_urls"] = alternateURLs
	}

	response.Success(c, result)
}

// ListVideos 获取视频列表
// GET /api/v1/videos
func (h *VideoHandler) ListVideos(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	category := c.Query("category")

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	videos, total, err := h.svc.ListVideos(page, pageSize, category)
	if err != nil {
		response.InternalError(c, "获取视频列表失败")
		return
	}

	response.SuccessPage(c, videos, total, page, pageSize)
}

// SearchVideos 搜索视频
// GET /api/v1/search
func (h *VideoHandler) SearchVideos(c *gin.Context) {
	keyword := c.Query("keyword")
	if keyword == "" {
		response.BadRequest(c, "搜索关键词不能为空")
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

	videos, total, err := h.svc.SearchVideos(keyword, page, pageSize)
	if err != nil {
		response.InternalError(c, "搜索失败")
		return
	}

	response.SuccessPage(c, videos, total, page, pageSize)
}

// GetCategories 获取分类列表
// GET /api/v1/categories
func (h *VideoHandler) GetCategories(c *gin.Context) {
	categories, err := h.svc.GetCategories()
	if err != nil {
		response.InternalError(c, "获取分类失败")
		return
	}

	response.Success(c, categories)
}

// GetRandom 获取随机推荐
// GET /api/v1/videos/random
func (h *VideoHandler) GetRandom(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	videos, err := h.svc.GetRandom(limit)
	if err != nil {
		response.InternalError(c, "获取推荐失败")
		return
	}

	response.Success(c, videos)
}

// GetLatest 获取最新视频
// GET /api/v1/videos/latest
func (h *VideoHandler) GetLatest(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	videos, err := h.svc.GetLatest(limit)
	if err != nil {
		response.InternalError(c, "获取最新视频失败")
		return
	}

	response.Success(c, videos)
}

// GetHot 获取热门视频
// GET /api/v1/videos/hot
func (h *VideoHandler) GetHot(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	videos, err := h.svc.GetHot(limit)
	if err != nil {
		response.InternalError(c, "获取热门视频失败")
		return
	}

	response.Success(c, videos)
}

// GetEpisodes 获取视频剧集
// GET /api/v1/videos/:id/episodes
func (h *VideoHandler) GetEpisodes(c *gin.Context) {
	id := c.Param("id")

	episodes, err := h.svc.GetEpisodes(id)
	if err != nil {
		response.InternalError(c, "获取剧集失败")
		return
	}

	response.Success(c, episodes)
}

// RecordWatch 记录观看
// POST /api/v1/videos/:id/watch
func (h *VideoHandler) RecordWatch(c *gin.Context) {
	id := c.Param("id")

	var req struct {
		Progress float64 `json:"progress"`
		Duration float64 `json:"duration"`
	}
	_ = c.ShouldBindJSON(&req)

	userID, _ := c.Get("user_id")
	uid, _ := userID.(string)

	if err := h.svc.RecordWatch(uid, id, req.Progress, req.Duration); err != nil {
		response.InternalError(c, "记录观看失败")
		return
	}

	response.SuccessWithMessage(c, "记录成功", nil)
}

// GetWatchHistory 获取观看历史
// GET /api/v1/user/history
func (h *VideoHandler) GetWatchHistory(c *gin.Context) {
	userID, _ := c.Get("user_id")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	histories, total, err := h.svc.GetWatchHistory(userID.(string), page, pageSize)
	if err != nil {
		response.InternalError(c, "获取观看历史失败")
		return
	}

	response.SuccessPage(c, histories, total, page, pageSize)
}

// GetSearchHot 获取热搜
// GET /api/v1/search/hot
func (h *VideoHandler) GetSearchHot(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit < 1 || limit > 50 {
		limit = 10
	}

	hots, err := h.svc.GetSearchHot(limit)
	if err != nil {
		response.InternalError(c, "获取热搜失败")
		return
	}

	response.Success(c, hots)
}

// enrichPlayLines 为播放线路附加资源站状态信息
// 如果监控服务未注入或无状态数据，则原样返回
func (h *VideoHandler) enrichPlayLines(playLines interface{}) interface{} {
	if h.monitor == nil || playLines == nil {
		return playLines
	}

	// 获取资源站状态映射（一次性获取，避免循环内重复查询）
	statusMap := h.monitor.GetStationStatusCompactMap()
	if len(statusMap) == 0 {
		return playLines
	}

	// 将 playLines 转为可遍历的结构
	// playLines 的实际类型是 model.PlayLinesJSON ([]model.PlayLineJSON)
	lines, ok := playLines.([]interface{})
	if !ok {
		// 尝试通过 JSON 序列化/反序列化处理
		return h.enrichPlayLinesViaJSON(playLines, statusMap)
	}

	// 为每条线路附加 station_status
	enriched := make([]interface{}, 0, len(lines))
	for _, line := range lines {
		if lineMap, ok := line.(map[string]interface{}); ok {
			// 复制原始数据
			enrichedLine := make(map[string]interface{})
			for k, v := range lineMap {
				enrichedLine[k] = v
			}

			// 查找对应的资源站状态
			if sourceName, ok := lineMap["source_name"].(string); ok {
				if status, found := statusMap[sourceName]; found {
					enrichedLine["station_status"] = gin.H{
						"is_alive":   status.IsAlive,
						"latency_ms": status.Latency,
						"speed_kbps": status.Speed,
						"weight":     status.Weight,
					}
				}
			}

			enriched = append(enriched, enrichedLine)
		} else {
			enriched = append(enriched, line)
		}
	}

	return enriched
}

// enrichPlayLinesViaJSON 通过 JSON 序列化/反序列化为播放线路附加状态
// 用于处理 model.PlayLinesJSON 等自定义类型的场景
func (h *VideoHandler) enrichPlayLinesViaJSON(playLines interface{}, statusMap map[string]service.StationStatusCompact) interface{} {
	// 使用 json 序列化再反序列化为 []map[string]interface{}
	data, err := json.Marshal(playLines)
	if err != nil {
		return playLines
	}

	var lines []map[string]interface{}
	if err := json.Unmarshal(data, &lines); err != nil {
		return playLines
	}

	for i, line := range lines {
		if sourceName, ok := line["source_name"].(string); ok {
			if status, found := statusMap[sourceName]; found {
				lines[i]["station_status"] = gin.H{
					"is_alive":   status.IsAlive,
					"latency_ms": status.Latency,
					"speed_kbps": status.Speed,
					"weight":     status.Weight,
				}
			}
		}
	}

	return lines
}
