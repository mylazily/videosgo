package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/service"
	"github.com/mylazily/videosgo/pkg/response"
)

// XHandler X.com 处理器
type XHandler struct {
	svc *service.XService
}

// NewXHandler 创建 X.com 处理器
func NewXHandler(svc *service.XService) *XHandler {
	return &XHandler{svc: svc}
}

// ListAccounts 账号列表
// GET /api/v1/x/accounts
func (h *XHandler) ListAccounts(c *gin.Context) {
	accounts, err := h.svc.GetAccounts()
	if err != nil {
		response.InternalError(c, "获取账号列表失败")
		return
	}
	response.Success(c, accounts)
}

// CreatePost 创建推文
// POST /api/v1/x/post
func (h *XHandler) CreatePost(c *gin.Context) {
	var req struct {
		AccountID string   `json:"account_id" binding:"required"`
		VideoID   string   `json:"video_id" binding:"required"`
		Text      string   `json:"text" binding:"required"`
		Hashtags  []string `json:"hashtags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	accountID, err := uuid.Parse(req.AccountID)
	if err != nil {
		response.BadRequest(c, "无效的账号 ID")
		return
	}

	videoID, err := uuid.Parse(req.VideoID)
	if err != nil {
		response.BadRequest(c, "无效的视频 ID")
		return
	}

	postLog, err := h.svc.CreatePost(accountID, videoID, req.Text, req.Hashtags)
	if err != nil {
		response.InternalError(c, "创建推文失败")
		return
	}

	response.Success(c, postLog)
}

// ListPosts 发布记录
// GET /api/v1/x/posts
func (h *XHandler) ListPosts(c *gin.Context) {
	posts, err := h.svc.GetPostLogs(50)
	if err != nil {
		response.InternalError(c, "获取发布记录失败")
		return
	}
	response.Success(c, posts)
}

// ProcessQueue 手动处理队列（管理员）
// POST /api/v1/x/process-queue
func (h *XHandler) ProcessQueue(c *gin.Context) {
	count, err := h.svc.ProcessQueue()
	if err != nil {
		response.InternalError(c, "处理队列失败")
		return
	}

	response.SuccessWithMessage(c, "队列处理完成", gin.H{
		"processed_count": count,
	})
}
