package handler

import (
	"videosgo/internal/service"
	"videosgo/pkg/response"

	"github.com/gin-gonic/gin"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// GetProfile 获取用户资料
func (h *UserHandler) GetProfile(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未登录")
		return
	}
	userID, ok := userIDVal.(string)
	if !ok {
		response.InternalError(c, "用户 ID 格式错误")
		return
	}

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.Success(c, user.ToResponse())
}

// UpdateProfile 更新用户资料
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Unauthorized(c, "用户未登录")
		return
	}
	userID, ok := userIDVal.(string)
	if !ok {
		response.InternalError(c, "用户 ID 格式错误")
		return
	}

	var req service.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.userService.UpdateProfile(userID, &req); err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "更新成功", nil)
}
