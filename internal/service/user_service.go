package service

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/mylazily/videosgo/internal/model"
	"github.com/mylazily/videosgo/internal/repository"
	"github.com/mylazily/videosgo/pkg/crypto"
	jwtpkg "github.com/mylazily/videosgo/pkg/jwt"
)

// UserService 用户服务
type UserService struct {
	repo   *repository.UserRepo
	jwtMgr *jwtpkg.JWTManager
}

// NewUserService 创建用户服务
func NewUserService(repo *repository.UserRepo, jwtMgr *jwtpkg.JWTManager) *UserService {
	return &UserService{
		repo:   repo,
		jwtMgr: jwtMgr,
	}
}

// Register 用户注册
func (s *UserService) Register(username, password string) (*model.User, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("用户名和密码不能为空")
	}
	if len(username) < 3 || len(username) > 50 {
		return nil, fmt.Errorf("用户名长度需在 3-50 个字符之间")
	}
	if len(password) < 6 {
		return nil, fmt.Errorf("密码长度不能少于 6 个字符")
	}

	// 检查用户名是否已存在
	exists, err := s.repo.ExistsByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("检查用户名失败: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("用户名已存在")
	}

	// 密码哈希
	hashedPassword, err := crypto.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	user := &model.User{
		Username: username,
		Password: hashedPassword,
		Status:   "active",
	}

	if err := s.repo.Create(user); err != nil {
		return nil, fmt.Errorf("创建用户失败: %w", err)
	}

	return user, nil
}

// Login 用户登录
func (s *UserService) Login(username, password string) (string, *model.User, error) {
	user, err := s.repo.GetByUsername(username)
	if err != nil {
		return "", nil, fmt.Errorf("用户名或密码错误")
	}

	if user.Status != "active" {
		return "", nil, fmt.Errorf("账号已被禁用")
	}

	if !crypto.CheckPassword(password, user.Password) {
		return "", nil, fmt.Errorf("用户名或密码错误")
	}

	token, err := s.jwtMgr.GenerateToken(user.ID.String(), user.Username, user.IsAdmin)
	if err != nil {
		return "", nil, fmt.Errorf("生成令牌失败: %w", err)
	}

	return token, user, nil
}

// GetUser 获取用户信息
func (s *UserService) GetUser(id string) (*model.User, error) {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %s", id)
	}
	return s.repo.GetByID(parsedID)
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(user *model.User) error {
	return s.repo.Update(user)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id string) error {
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return fmt.Errorf("invalid UUID: %s", id)
	}
	return s.repo.Delete(parsedID)
}

// ListUsers 获取用户列表（管理员）
func (s *UserService) ListUsers(page, pageSize int) ([]model.User, int64, error) {
	return s.repo.List(page, pageSize)
}

// RefreshToken 刷新令牌
func (s *UserService) RefreshToken(userID string) (string, error) {
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return "", fmt.Errorf("invalid UUID: %s", userID)
	}
	user, err := s.repo.GetByID(parsedID)
	if err != nil {
		return "", fmt.Errorf("用户不存在")
	}
	return s.jwtMgr.GenerateToken(user.ID.String(), user.Username, user.IsAdmin)
}

// ChangePassword 修改密码
func (s *UserService) ChangePassword(userID, oldPassword, newPassword string) error {
	// 验证新密码长度
	if len(newPassword) < 6 {
		return fmt.Errorf("新密码长度不能少于 6 个字符")
	}

	// 获取用户
	parsedID, err := uuid.Parse(userID)
	if err != nil {
		return fmt.Errorf("invalid UUID: %s", userID)
	}
	user, err := s.repo.GetByID(parsedID)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}

	// 验证旧密码
	if !crypto.CheckPassword(oldPassword, user.Password) {
		return fmt.Errorf("旧密码错误")
	}

	// 哈希新密码
	hashedPassword, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新密码
	user.Password = hashedPassword
	return s.repo.Update(user)
}
