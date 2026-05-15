package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/service"
	"github.com/mylazily/videosgo/pkg/response"
)

// CommentHandler 评论处理器
type CommentHandler struct {
	svc *service.CommentService
}

// NewCommentHandler 创建评论处理器
func NewCommentHandler(svc *service.CommentService) *CommentHandler {
	return &CommentHandler{svc: svc}
}

// CreateComment 创建评论
// POST /api/v1/videos/:id/comments
func (h *CommentHandler) CreateComment(c *gin.Context) {
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的视频 ID")
		return
	}

	var req struct {
		Content  string `json:"content" binding:"required"`
		ParentID uint   `json:"parent_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("user_id")

	comment := &model.Comment{
		VideoID:  uint(videoID),
		UserID:   userID.(uint),
		Content:  req.Content,
		ParentID: req.ParentID,
	}

	if err := h.svc.CreateComment(comment); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, comment)
}

// DeleteComment 删除评论
// DELETE /api/v1/comments/:id
func (h *CommentHandler) DeleteComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论 ID")
		return
	}

	userID, _ := c.Get("user_id")
	isAdmin, _ := c.Get("is_admin")

	if err := h.svc.DeleteComment(uint(id), userID.(uint), isAdmin.(bool)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// ListComments 获取视频评论列表
// GET /api/v1/videos/:id/comments
func (h *CommentHandler) ListComments(c *gin.Context) {
	videoID, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的视频 ID")
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

	comments, total, err := h.svc.ListComments(uint(videoID), page, pageSize)
	if err != nil {
		response.InternalError(c, "获取评论列表失败")
		return
	}

	response.SuccessPage(c, comments, total, page, pageSize)
}

// ListReplies 获取回复列表
// GET /api/v1/comments/:id/replies
func (h *CommentHandler) ListReplies(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论 ID")
		return
	}

	replies, err := h.svc.ListReplies(uint(id))
	if err != nil {
		response.InternalError(c, "获取回复列表失败")
		return
	}

	response.Success(c, replies)
}

// LikeComment 点赞评论
// POST /api/v1/comments/:id/like
func (h *CommentHandler) LikeComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论 ID")
		return
	}

	userID, _ := c.Get("user_id")

	if err := h.svc.LikeComment(uint(id), userID.(uint)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "点赞成功", nil)
}

// UnlikeComment 取消点赞
// DELETE /api/v1/comments/:id/like
func (h *CommentHandler) UnlikeComment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "无效的评论 ID")
		return
	}

	userID, _ := c.Get("user_id")

	if err := h.svc.UnlikeComment(uint(id), userID.(uint)); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "取消点赞成功", nil)
}
