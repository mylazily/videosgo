package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"videosgo/internal/service"
	"videosgo/pkg/response"
)

// AdRewardHandler 广告金币处理器
type AdRewardHandler struct {
	svc *service.AdRewardService
}

// NewAdRewardHandler 创建广告金币处理器
func NewAdRewardHandler(svc *service.AdRewardService) *AdRewardHandler {
	return &AdRewardHandler{svc: svc}
}

// ListTasks 任务列表
// GET /api/v1/reward/tasks
func (h *AdRewardHandler) ListTasks(c *gin.Context) {
	tasks, err := h.svc.GetTasks()
	if err != nil {
		response.InternalError(c, "获取任务列表失败")
		return
	}
	response.Success(c, tasks)
}

// CompleteTask 完成任务
// POST /api/v1/reward/complete
func (h *AdRewardHandler) CompleteTask(c *gin.Context) {
	var req struct {
		FingerprintID string `json:"fingerprint_id" binding:"required"`
		TaskID        string `json:"task_id" binding:"required"`
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

	taskID, err := uuid.Parse(req.TaskID)
	if err != nil {
		response.BadRequest(c, "无效的任务 ID")
		return
	}

	tx, err := h.svc.CompleteTask(fingerprintID, taskID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "任务完成，奖励已发放", tx)
}

// GetBalance 金币余额
// GET /api/v1/reward/balance
func (h *AdRewardHandler) GetBalance(c *gin.Context) {
	fingerprintIDStr := c.Query("fingerprint_id")
	if fingerprintIDStr == "" {
		response.BadRequest(c, "设备指纹 ID 不能为空")
		return
	}

	fingerprintID, err := uuid.Parse(fingerprintIDStr)
	if err != nil {
		response.BadRequest(c, "无效的设备指纹 ID")
		return
	}

	balance, err := h.svc.GetBalance(fingerprintID)
	if err != nil {
		response.InternalError(c, "获取余额失败")
		return
	}

	response.Success(c, gin.H{
		"balance": balance,
	})
}

// GetHistory 金币流水
// GET /api/v1/reward/history
func (h *AdRewardHandler) GetHistory(c *gin.Context) {
	fingerprintIDStr := c.Query("fingerprint_id")
	if fingerprintIDStr == "" {
		response.BadRequest(c, "设备指纹 ID 不能为空")
		return
	}

	fingerprintID, err := uuid.Parse(fingerprintIDStr)
	if err != nil {
		response.BadRequest(c, "无效的设备指纹 ID")
		return
	}

	history, err := h.svc.GetTransactionHistory(fingerprintID, 50)
	if err != nil {
		response.InternalError(c, "获取流水失败")
		return
	}

	response.Success(c, history)
}

// UnlockVideo 金币解锁视频
// POST /api/v1/reward/unlock
func (h *AdRewardHandler) UnlockVideo(c *gin.Context) {
	var req struct {
		FingerprintID string `json:"fingerprint_id" binding:"required"`
		VideoID       string `json:"video_id" binding:"required"`
		Cost          int64  `json:"cost" binding:"required"`
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

	tx, err := h.svc.UnlockVideoWithCoins(fingerprintID, videoID, req.Cost)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "解锁成功", tx)
}

// DailyCheckIn 每日签到
// POST /api/v1/reward/checkin
func (h *AdRewardHandler) DailyCheckIn(c *gin.Context) {
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

	tx, err := h.svc.DailyCheckIn(fingerprintID)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "签到成功", tx)
}

// GetDashboard 奖励面板
// GET /api/v1/reward/dashboard
func (h *AdRewardHandler) GetDashboard(c *gin.Context) {
	fingerprintIDStr := c.Query("fingerprint_id")
	if fingerprintIDStr == "" {
		response.BadRequest(c, "设备指纹 ID 不能为空")
		return
	}

	fingerprintID, err := uuid.Parse(fingerprintIDStr)
	if err != nil {
		response.BadRequest(c, "无效的设备指纹 ID")
		return
	}

	dashboard, err := h.svc.GetRewardDashboard(fingerprintID)
	if err != nil {
		response.InternalError(c, "获取奖励面板失败")
		return
	}

	response.Success(c, dashboard)
}
