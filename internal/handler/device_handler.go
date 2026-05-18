package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"videosgo/internal/service"
	"videosgo/pkg/response"
)

// DeviceHandler 设备指纹处理器
type DeviceHandler struct {
	svc *service.DeviceService
}

// NewDeviceHandler 创建设备指纹处理器
func NewDeviceHandler(svc *service.DeviceService) *DeviceHandler {
	return &DeviceHandler{svc: svc}
}

// RegisterDevice 注册设备
// POST /api/v1/device/register
func (h *DeviceHandler) RegisterDevice(c *gin.Context) {
	var req struct {
		Hash       string `json:"fingerprint_hash" binding:"required"`
		UserAgent  string `json:"user_agent"`
		Screen     string `json:"screen_resolution"`
		Language   string `json:"language"`
		Timezone   string `json:"timezone"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	device, err := h.svc.AuthenticateDevice(req.Hash, req.UserAgent, req.Screen, req.Language, req.Timezone)
	if err != nil {
		response.InternalError(c, "设备注册失败")
		return
	}

	response.Success(c, device)
}

// GetDeviceProfile 获取设备信息（硬币余额等）
// GET /api/v1/device/profile?fingerprint_id=xxx
func (h *DeviceHandler) GetDeviceProfile(c *gin.Context) {
	idStr := c.Query("fingerprint_id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(c, "无效的设备指纹 ID")
		return
	}

	// 获取设备信息
	device, err := h.svc.GetDevice(id)
	if err != nil {
		response.NotFound(c, "设备不存在")
		return
	}

	// 获取硬币余额
	balance, err := h.svc.GetCoinBalance(id)
	if err != nil {
		response.InternalError(c, "获取硬币余额失败")
		return
	}

	response.Success(c, gin.H{
		"device":  device,
		"balance": balance,
	})
}

// UnlockVideo 解锁视频
// POST /api/v1/device/unlock
func (h *DeviceHandler) UnlockVideo(c *gin.Context) {
	var req struct {
		FingerprintID string `json:"fingerprint_id" binding:"required"`
		VideoID       string `json:"video_id" binding:"required"`
		UnlockType    string `json:"unlock_type" binding:"required"`
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

	if err := h.svc.UnlockVideo(fingerprintID, videoID, req.UnlockType); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "解锁成功", nil)
}

// CheckVideoUnlocked 检查视频是否已解锁
// GET /api/v1/device/check/:videoId?fingerprint_id=xxx
func (h *DeviceHandler) CheckVideoUnlocked(c *gin.Context) {
	videoIDStr := c.Param("videoId")
	videoID, err := uuid.Parse(videoIDStr)
	if err != nil {
		response.BadRequest(c, "无效的视频 ID")
		return
	}

	fingerprintIDStr := c.Query("fingerprint_id")
	fingerprintID, err := uuid.Parse(fingerprintIDStr)
	if err != nil {
		response.BadRequest(c, "无效的设备指纹 ID")
		return
	}

	unlocked, err := h.svc.CheckAccess(fingerprintID, videoID)
	if err != nil {
		response.InternalError(c, "检查解锁状态失败")
		return
	}

	response.Success(c, gin.H{
		"unlocked": unlocked,
	})
}
