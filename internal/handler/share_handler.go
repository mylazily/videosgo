package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/service"
	"github.com/mylazily/videosgo/pkg/response"
)

// ShareHandler 分享裂变处理器
type ShareHandler struct {
	svc *service.ShareService
}

// NewShareHandler 创建分享裂变处理器
func NewShareHandler(svc *service.ShareService) *ShareHandler {
	return &ShareHandler{svc: svc}
}

// CreateShareLink 创建分享链接
// POST /api/v1/share/create
func (h *ShareHandler) CreateShareLink(c *gin.Context) {
	var req struct {
		FingerprintID string `json:"fingerprint_id" binding:"required"`
		VideoID       string `json:"video_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	fingerprintID, err := uuid.Parse(req.FingerprintID)
	if err != nil {
		response.BadRequest(c, "无效的设备指纹 ID")
		return
	}

	videoID, err := uuid.Parse(req.VideoID)
	if err != nil {
		response.BadRequest(c, "无效的视频 ID")
		return
	}

	link, err := h.svc.GenerateShareLink(fingerprintID, videoID)
	if err != nil {
		response.InternalError(c, "创建分享链接失败")
		return
	}

	response.Success(c, link)
}

// GetShareLink 获取分享信息
// GET /api/v1/share/:code
func (h *ShareHandler) GetShareLink(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "分享码不能为空")
		return
	}

	link, err := h.svc.GetShareLink(code)
	if err != nil {
		response.NotFound(c, "分享链接不存在")
		return
	}

	response.Success(c, link)
}

// RecordShareClick 记录分享点击
// POST /api/v1/share/:code/click
func (h *ShareHandler) RecordShareClick(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		response.BadRequest(c, "分享码不能为空")
		return
	}

	var req struct {
		FingerprintID string `json:"fingerprint_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	fingerprintID, err := uuid.Parse(req.FingerprintID)
	if err != nil {
		response.BadRequest(c, "无效的设备指纹 ID")
		return
	}

	ip := c.ClientIP()
	ua := c.GetHeader("User-Agent")

	link, err := h.svc.ProcessShareClick(code, fingerprintID, ip, ua)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, link)
}
