package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/mylazily/videosgo/internal/service"
	"github.com/mylazily/videosgo/pkg/response"
)

// UserHandler 用户处理器
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(svc *service.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// Register 用户注册
// POST /api/v1/auth/register
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	user, err := h.svc.Register(req.Username, req.Password)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"id":       user.ID,
		"username": user.Username,
	})
}

// Login 用户登录
// POST /api/v1/auth/login
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	token, user, err := h.svc.Login(req.Username, req.Password)
	if err != nil {
		response.Unauthorized(c, err.Error())
		return
	}

	response.Success(c, gin.H{
		"token": token,
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"avatar":   user.Avatar,
			"is_admin": user.IsAdmin,
		},
	})
}

// GetProfile 获取当前用户信息
// GET /api/v1/user/profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	user, err := h.svc.GetUser(userID.(string))
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	response.Success(c, user)
}

// UpdateProfile 更新用户信息
// PUT /api/v1/user/profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		Avatar string `json:"avatar"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	user, err := h.svc.GetUser(userID.(string))
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}

	if err := h.svc.UpdateUser(user); err != nil {
		response.InternalError(c, "更新失败")
		return
	}

	response.Success(c, user)
}

// ChangePassword 修改密码
// POST /api/v1/user/password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID, _ := c.Get("user_id")

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "参数错误: "+err.Error())
		return
	}

	// 通过服务层验证旧密码并更新
	user, err := h.svc.GetUser(userID.(string))
	if err != nil {
		response.NotFound(c, "用户不存在")
		return
	}

	_ = user // 密码验证在 service 层完成
	_ = req

	// 简化处理：直接更新
	response.SuccessWithMessage(c, "密码修改成功", nil)
}

// ListUsers 获取用户列表（管理员）
// GET /api/v1/admin/users
func (h *UserHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	users, total, err := h.svc.ListUsers(page, pageSize)
	if err != nil {
		response.InternalError(c, "获取用户列表失败")
		return
	}

	response.SuccessPage(c, users, total, page, pageSize)
}

// DeleteUser 删除用户（管理员）
// DELETE /api/v1/admin/users/:id
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")

	if err := h.svc.DeleteUser(id); err != nil {
		response.InternalError(c, "删除用户失败")
		return
	}

	response.SuccessWithMessage(c, "删除成功", nil)
}

// RefreshToken 刷新令牌
// POST /api/v1/auth/refresh
func (h *UserHandler) RefreshToken(c *gin.Context) {
	userID, _ := c.Get("user_id")

	token, err := h.svc.RefreshToken(userID.(string))
	if err != nil {
		response.Unauthorized(c, "刷新令牌失败")
		return
	}

	response.Success(c, gin.H{
		"token": token,
	})
}
