package service

import (
	"time"

	"videosgo/internal/model"
	"videosgo/internal/repository"
)

// UserService 用户服务
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetProfile 获取用户资料
func (s *UserService) GetProfile(userID string) (*model.User, error) {
	return s.userRepo.GetByID(userID)
}

// UpdateProfileRequest 更新资料请求
type UpdateProfileRequest struct {
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
}

// UpdateProfile 更新用户资料
func (s *UserService) UpdateProfile(userID string, req *UpdateProfileRequest) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Avatar != "" {
		user.Avatar = req.Avatar
	}
	user.UpdatedAt = time.Now()

	return s.userRepo.Update(user)
}
